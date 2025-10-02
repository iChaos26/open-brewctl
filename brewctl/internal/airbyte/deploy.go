package airbyte

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func Deploy() error {
	fmt.Println("📦 Deploying Airbyte...")

	// Add Airbyte Helm repo
	cmd := exec.Command("helm", "repo", "add", "airbyte", "https://airbytehq.github.io/helm-charts")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add airbyte repo: %v", err)
	}

	// Update Helm repos
	cmd = exec.Command("helm", "repo", "update")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update helm repos: %v", err)
	}

	// Deploy Airbyte
	cmd = exec.Command("helm", "install", "airbyte", "airbyte/airbyte",
		"--namespace", "default",
		"--set", "global.service.type=NodePort",
		"--set", "server.service.nodePorts.api=8000",
		"--set", "worker.enabled=true",
		"--set", "ingress.enabled=true",
		"--set", "ingress.className=nginx",
		"--wait",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to deploy Airbyte: %v", err)
	}

	// Wait for Airbyte to be ready
	time.Sleep(60 * time.Second)

	fmt.Println("✅ Airbyte deployed successfully")
	return nil
}
