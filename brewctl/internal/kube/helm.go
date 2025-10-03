package kube

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func DeployMongoDB() error {
	fmt.Println("📦 Implantando MongoDB Community Edition...")
	return implantarMongoDBDireto()
}

func implantarMongoDBDireto() error {
	fmt.Println("📦 Executando implantação direta do MongoDB...")

	// Limpar recursos existentes
	cmd := exec.Command("sh", "-c", `
        kubectl delete deployment mongodb --ignore-not-found=true;
        kubectl delete service mongodb --ignore-not-found=true;
        sleep 2;
    `)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	// Aplicar deployment do MongoDB
	cmd = exec.Command("kubectl", "apply", "-f", "-")
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
    nodePort: 30017
  selector:
    app: mongodb
`)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("falha na implantação do MongoDB: %v", err)
	}

	fmt.Println("✅ MongoDB implantado com sucesso!")
	return nil
}
