package kube

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func CreateKindCluster() error {
	fmt.Println("🔧 Creating Kind Kubernetes cluster...")

	// ✅ USAR CONFIGURAÇÃO ALTERNATIVA (sem portas 80/443)
	kindConfig := `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: brewctl-cluster
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 8000
    hostPort: 8000
    protocol: TCP
  - containerPort: 3000
    hostPort: 3000
    protocol: TCP  
  - containerPort: 9090
    hostPort: 9090
    protocol: TCP
  - containerPort: 27017
    hostPort: 27017
    protocol: TCP`

	configPath := filepath.Join(os.TempDir(), "kind-config-brewctl.yaml")
	if err := os.WriteFile(configPath, []byte(kindConfig), 0644); err != nil {
		return fmt.Errorf("failed to write kind config: %v", err)
	}
	defer os.Remove(configPath)

	// Resto do código permanece igual...
	cmd := exec.Command("kind", "create", "cluster", "--config", configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create kind cluster: %v", err)
	}

	// Wait for cluster to be ready
	fmt.Println("⏳ Waiting for cluster to be ready...")
	time.Sleep(45 * time.Second)

	// Verify cluster
	if err := CheckClusterStatus(); err != nil {
		return fmt.Errorf("cluster verification failed: %v", err)
	}

	fmt.Println("✅ Kind cluster created and verified")
	return nil
}

func CheckClusterStatus() error {
	// Check cluster info
	cmd := exec.Command("kubectl", "cluster-info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kubernetes cluster not accessible: %v", err)
	}

	// Check nodes
	cmd = exec.Command("kubectl", "get", "nodes", "-o", "wide")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to get nodes: %v", err)
	}

	return nil
}
