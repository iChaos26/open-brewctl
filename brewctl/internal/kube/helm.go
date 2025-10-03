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
func DeployMongoDB() error {
	fmt.Println("📦 Deploying MongoDB with Helm...")

	// Add Bitnami repo
	if err := runHelmCommand("repo", "add", "bitnami", "https://charts.bitnami.com/bitnami"); err != nil {
		return fmt.Errorf("failed to add bitnami repo: %v", err)
	}
	if err := runHelmCommand("repo", "update"); err != nil {
		return fmt.Errorf("failed to update helm repos: %v", err)
	}

	// Limpar instalação existente
	runHelmCommand("uninstall", "mongodb")
	time.Sleep(5 * time.Second)

	// Tentar versões estáveis e validadas do Bitnami
	validImages := []struct {
		tag         string
		setPlatform bool
	}{
		// Tags Bitnami validadas para amd64
		{tag: "6.0.4-debian-11-r30", setPlatform: true},
		{tag: "5.0.8-debian-11-r23", setPlatform: true},
		// Tentativa com imagem oficial do MongoDB Community
		{tag: "6.0.5", setPlatform: false},
	}

	for i, img := range validImages {
		fmt.Printf("🔄 Attempt %d/%d: Deploying MongoDB with tag %s\n", i+1, len(validImages), img.tag)

		args := []string{
			"install", "mongodb", "bitnami/mongodb",
			"--namespace", "default",
			"--set", "auth.enabled=false",
			"--set", "persistence.enabled=true",
			"--set", "service.type=NodePort",
			"--set", "service.nodePort=27017",
			"--set", "replicaSet.enabled=false",
			"--set", "image.tag=" + img.tag,
			"--set", "image.pullPolicy=Always",
			"--wait",
			"--timeout", "10m",
		}

		// Forçar plataforma AMD64 se necessário (para Mac M1)
		if img.setPlatform {
			args = append(args, "--set", "image.platform=linux/amd64")
		}

		cmd := exec.Command("helm", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️ Attempt with tag %s failed: %v\n", img.tag, err)
			// Limpar antes da próxima tentativa
			runHelmCommand("uninstall", "mongodb")
			time.Sleep(10 * time.Second)
			continue
		}

		// Verificar se o pod ficou Ready
		fmt.Println("⏳ Waiting for MongoDB to be ready...")
		time.Sleep(30 * time.Second)

		if checkPodStatus() {
			fmt.Printf("✅ MongoDB successfully deployed with tag %s\n", img.tag)
			return nil
		}
		runHelmCommand("uninstall", "mongodb")
	}
	return fmt.Errorf("❌ failed to deploy MongoDB after all attempts")
}

func checkPodStatus() bool {
	// Verificar status e eventos do pod
	cmd := exec.Command("kubectl", "get", "pods", "-l", "app.kubernetes.io/name=mongodb", "-o", "wide")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	// Comando para ver eventos detalhados do pod
	cmd = exec.Command("kubectl", "describe", "pod", "-l", "app.kubernetes.io/name=mongodb")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	// Verificar se o container está Ready
	cmd = exec.Command("kubectl", "get", "pod", "-l", "app.kubernetes.io/name=mongodb", "-o", "jsonpath={.items[0].status.containerStatuses[0].ready}")
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) == "true"
}
