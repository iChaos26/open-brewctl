package monitoring

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func DeployPrometheus() error {
	fmt.Println("📊 Deploying Prometheus...")

	// Add Prometheus Helm repo
	cmd := exec.Command("helm", "repo", "add", "prometheus-community", "https://prometheus-community.github.io/helm-charts")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add prometheus repo: %v", err)
	}

	// Update Helm repos
	cmd = exec.Command("helm", "repo", "update")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update helm repos: %v", err)
	}

	// Deploy Prometheus
	cmd = exec.Command("helm", "install", "prometheus", "prometheus-community/prometheus",
		"--namespace", "default",
		"--set", "server.service.type=NodePort",
		"--set", "server.service.nodePort=9090",
		"--set", "alertmanager.enabled=false",
		"--set", "pushgateway.enabled=false",
		"--set", "nodeExporter.enabled=false",
		"--wait",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to deploy Prometheus: %v", err)
	}

	time.Sleep(30 * time.Second)
	fmt.Println("✅ Prometheus deployed successfully")
	return nil
}
