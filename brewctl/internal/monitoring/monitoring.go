package monitoring

import "fmt"

// ✅ ADICIONAR: Função Deploy que integra Prometheus + Grafana
func Deploy() error {
	fmt.Println("📊 Deploying monitoring stack...")

	if err := DeployPrometheus(); err != nil {
		return fmt.Errorf("failed to deploy Prometheus: %v", err)
	}

	if err := DeployGrafana(); err != nil {
		return fmt.Errorf("failed to deploy Grafana: %v", err)
	}

	fmt.Println("✅ Monitoring stack deployed successfully")
	return nil
}
