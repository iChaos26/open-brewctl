package kube

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func DeployMongoDB() error {
	fmt.Println("📦 Deploying MongoDB with Helm...")

	// Add Bitnami repo
	if err := runHelmCommand("repo", "add", "bitnami", "https://charts.bitnami.com/bitnami"); err != nil {
		return fmt.Errorf("failed to add bitnami repo: %v", err)
	}

	if err := runHelmCommand("repo", "update"); err != nil {
		return fmt.Errorf("failed to update helm repos: %v", err)
	}

	// Deploy MongoDB
	cmd := exec.Command("helm", "install", "mongodb", "bitnami/mongodb",
		"--namespace", "default",
		"--set", "auth.enabled=false",
		"--set", "persistence.enabled=true",
		"--set", "service.type=NodePort",
		"--set", "service.nodePort=27017",
		"--set", "replicaSet.enabled=false",
		"--wait",
		"--timeout", "10m",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to deploy MongoDB: %v", err)
	}

	// Wait for MongoDB to be ready
	fmt.Println("⏳ Waiting for MongoDB to be ready...")
	time.Sleep(60 * time.Second)

	// Verify MongoDB deployment
	cmd = exec.Command("kubectl", "get", "pods", "-l", "app.kubernetes.io/name=mongodb", "-o", "jsonpath={.items[0].status.phase}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check MongoDB pod status: %v", err)
	}

	if string(output) != "Running" {
		return fmt.Errorf("MongoDB pod is not running: %s", string(output))
	}

	fmt.Println("✅ MongoDB deployed successfully")
	return nil
}

func runHelmCommand(args ...string) error {
	cmd := exec.Command("helm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
