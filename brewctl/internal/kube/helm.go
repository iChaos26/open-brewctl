package kube

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func runHelmCommand(args ...string) error {
	cmd := exec.Command("helm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func kubectlDeletePVC(releaseName string) {
	cmd := exec.Command("kubectl", "delete", "pvc", "-l", "app.kubernetes.io/instance="+releaseName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func isMongoDBReady(label string) bool {
	maxRetries := 12
	for i := 0; i < maxRetries; i++ {
		// Check pod status and readiness
		cmd := exec.Command("kubectl", "get", "pod", "-l", label, "-o", "jsonpath={.items[0].status.containerStatuses[0].ready}")
		output, err := cmd.Output()
		if err == nil && strings.TrimSpace(string(output)) == "true" {
			return true
		}
		time.Sleep(10 * time.Second)
	}
	return false
}

func DeployMongoDB() error {
	fmt.Println("📦 Iniciando deploy do MongoDB...")

	// Adicionar e atualizar repositório Bitnami
	if err := runHelmCommand("repo", "add", "bitnami", "https://charts.bitnami.com/bitnami"); err != nil {
		return fmt.Errorf("falha ao adicionar repositório bitnami: %v", err)
	}
	if err := runHelmCommand("repo", "update"); err != nil {
		return fmt.Errorf("falha ao atualizar repositórios helm: %v", err)
	}

	// Limpar instalação existente
	fmt.Println("🧹 Limpando instalações existentes do MongoDB...")
	runHelmCommand("uninstall", "mongodb")
	time.Sleep(5 * time.Second)

	// Tentativa PRINCIPAL com a tag específica do Bitnami
	fmt.Println("🔄 Tentativa 1/2: Deployando com a tag Bitnami '4.0.10-debian-9-r47'")

	args := []string{
		"install", "mongodb", "bitnami/mongodb",
		"--namespace", "default",
		"--set", "auth.enabled=false",
		"--set", "persistence.enabled=true",
		"--set", "service.type=NodePort",
		"--set", "service.nodePort=27017",
		"--set", "replicaSet.enabled=false",
		"--set", "image.tag=4.0.10-debian-9-r47",
		"--set", "image.pullPolicy=Always",
		"--wait",
		"--timeout", "10m",
	}

	cmd := exec.Command("helm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠️ Falha no deploy com Bitnami: %v\n", err)
		fmt.Println("🚀 Acionando fallback: MongoDB Community Edition...")

		// FALLBACK: Implantação direta com imagem oficial
		return deployMongoDBDirect()
	}

	// Verificar se o pod do Bitnami ficou Ready
	fmt.Println("⏳ Verificando status do MongoDB Bitnami...")
	if isMongoDBReady("app.kubernetes.io/name=mongodb") {
		fmt.Println("✅ MongoDB Bitnami deployado com sucesso!")
		return nil
	}

	fmt.Println("⚠️ Pod Bitnami não ficou pronto, acionando fallback...")
	runHelmCommand("uninstall", "mongodb")
	kubectlDeletePVC("mongodb")
	return deployMongoDBDirect()
}

// Fallback: Deploy direto usando a imagem oficial do MongoDB
func deployMongoDBDirect() error {
	fmt.Println("📦 Executando fallback: Deploy direto do MongoDB Community...")

	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: mongodb
  labels:
    app: mongodb
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mongodb
  template:
    metadata:
      labels:
        app: mongodb
    spec:
      containers:
      - name: mongodb
        image: mongo:6.0.5
        ports:
        - containerPort: 27017
        env:
        - name: MONGO_INITDB_DATABASE
          value: breweries_db
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: mongodb
spec:
  type: NodePort
  ports:
  - port: 27017
    targetPort: 27017
    nodePort: 27017
  selector:
    app: mongodb`)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("falha no deploy direto do MongoDB: %v", err)
	}

	// Verificar se o pod oficial ficou Ready
	fmt.Println("⏳ Verificando status do MongoDB Community...")
	time.Sleep(10 * time.Second)

	if isMongoDBReady("app=mongodb") {
		fmt.Println("✅ MongoDB Community Edition deployado com sucesso!")
		return nil
	}

	return fmt.Errorf("timeout: MongoDB Community Edition não ficou pronto")
}
