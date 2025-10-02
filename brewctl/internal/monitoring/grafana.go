package monitoring

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func DeployGrafana() error {
	fmt.Println("📈 Deploying Grafana...")

	// Add Grafana Helm repo
	cmd := exec.Command("helm", "repo", "add", "grafana", "https://grafana.github.io/helm-charts")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add grafana repo: %v", err)
	}

	// Update Helm repos
	cmd = exec.Command("helm", "repo", "update")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update helm repos: %v", err)
	}

	// Deploy Grafana
	cmd = exec.Command("helm", "install", "grafana", "grafana/grafana",
		"--namespace", "default",
		"--set", "service.type=NodePort",
		"--set", "service.nodePort=3000",
		"--set", "adminPassword=admin",
		"--set", "persistence.enabled=true",
		"--set", "persistence.size=10Gi",
		"--wait",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to deploy Grafana: %v", err)
	}

	time.Sleep(30 * time.Second)
	fmt.Println("✅ Grafana deployed successfully")
	return nil
}
