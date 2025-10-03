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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add airbyte repo: %v", err)
	}

	// Update Helm repos
	cmd = exec.Command("helm", "repo", "update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update helm repos: %v", err)
	}

	// Deploy Airbyte with optimized settings
	cmd = exec.Command("helm", "upgrade", "--install", "airbyte", "airbyte/airbyte",
		"--namespace", "default",
		"--set", "global.service.type=NodePort",
		"--set", "server.service.nodePorts.api=8000",
		"--set", "worker.enabled=true",
		"--set", "bootloader.enabled=true",
		"--set", "ingress.enabled=false",
		"--set", "resources.server.requests.memory=1Gi",
		"--set", "resources.server.requests.cpu=500m",
		"--wait",
		"--timeout", "15m",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to deploy Airbyte: %v", err)
	}

	// Wait for Airbyte to be ready
	fmt.Println("⏳ Waiting for Airbyte to be ready...")

	// Wait for Airbyte pods to be ready
	cmd = exec.Command("kubectl", "wait", "--for=condition=Ready", "pod", "-l", "app.kubernetes.io/name=airbyte", "--timeout=600s")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠️ Some pods took longer than expected, continuing anyway...\n")
	}

	// Additional wait for internal services to be stable
	time.Sleep(30 * time.Second)

	fmt.Println("✅ Airbyte deployed successfully")
	return nil
}
