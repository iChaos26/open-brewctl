package main

import (
	"fmt"
	"log"

	"brewctl/internal/airbyte"
	"brewctl/internal/kube"
	"brewctl/internal/mongodb"
	"brewctl/internal/monitoring"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
)

var rootCmd = &cobra.Command{
	Use:   "brewctl",
	Short: "Brewctl - Complete Data Pipeline for Breweries",
	Long: `A complete CLI tool for managing breweries data pipeline inspired by abctl.
Features:
• Kubernetes cluster management with Kind
• Airbyte data pipelines  
• MongoDB with aggregation pipelines
• Monitoring with Prometheus/Grafana
• Bronze/Silver/Gold data layers`,
}

var clusterInitCmd = &cobra.Command{
	Use:   "cluster-init",
	Short: "Initialize complete local Kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Initializing Breweries Data Cluster...")

		if err := kube.CreateKindCluster(); err != nil {
			log.Fatalf("❌ Failed to create Kind cluster: %v", err)
		}

		if err := airbyte.Deploy(); err != nil {
			log.Fatalf("❌ Failed to deploy Airbyte: %v", err)
		}

		if err := kube.DeployMongoDB(); err != nil {
			log.Fatalf("❌ Failed to deploy MongoDB: %v", err)
		}

		// ✅ CORREÇÃO: Usar monitoring.Deploy() em vez das funções individuais
		if err := monitoring.Deploy(); err != nil {
			log.Fatalf("❌ Failed to deploy monitoring stack: %v", err)
		}

		fmt.Println("✅ Cluster initialization completed!")
		fmt.Println("🌐 Airbyte: http://localhost:8000")
		fmt.Println("📊 Grafana: http://localhost:3000 (admin/admin)")
		fmt.Println("📈 Prometheus: http://localhost:9090")
		fmt.Println("🍃 MongoDB: localhost:27017")
	},
}

var deployConnectionsCmd = &cobra.Command{
	Use:   "deploy-connections",
	Short: "Deploy Airbyte source and destination connections",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔗 Deploying Airbyte connections...")

		client := airbyte.NewAirbyteClient("http://localhost:8000")
		if err := client.SetupConnections(); err != nil {
			log.Fatalf("❌ Failed to deploy connections: %v", err)
		}

		fmt.Println("✅ Airbyte connections deployed successfully!")
		fmt.Println("💡 Manual step: Trigger sync in Airbyte UI at http://localhost:8000")
	},
}

var runAggregationsCmd = &cobra.Command{
	Use:   "run-aggregations",
	Short: "Run MongoDB aggregation pipelines (Silver → Gold layers)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔄 Running MongoDB aggregation pipelines...")

		aggService, err := mongodb.NewAggregationService("mongodb://localhost:27017")
		if err != nil {
			log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
		}
		defer aggService.Close()

		// Run Silver Layer
		if err := aggService.RunSilverLayerAggregation(); err != nil {
			log.Fatalf("❌ Silver layer aggregation failed: %v", err)
		}

		// Run Gold Layer
		if err := aggService.RunGoldLayerAggregation(); err != nil {
			log.Fatalf("❌ Gold layer aggregation failed: %v", err)
		}

		// Show results
		fmt.Println("📊 Aggregation Results:")

		// Top states
		topStates, err := aggService.GetTopStates(5)
		if err != nil {
			log.Printf("⚠️ Failed to get top states: %v", err)
		} else {
			fmt.Println("🏆 Top 5 States by Brewery Count:")
			for i, state := range topStates {
				fmt.Printf("  %d. %s: %d breweries\n", i+1, state["state"], state["total_breweries"])
			}
		}

		// Brewery types
		typeDist, err := aggService.GetBreweryTypesDistribution()
		if err != nil {
			log.Printf("⚠️ Failed to get type distribution: %v", err)
		} else {
			fmt.Println("🍻 Brewery Type Distribution:")
			for _, dist := range typeDist {
				fmt.Printf("  • %s: %d (across %d states)\n",
					dist["brewery_type"], dist["count"], dist["states_covered"])
			}
		}

		fmt.Println("✅ All aggregations completed successfully!")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check cluster and services status",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Checking cluster status...")

		if err := kube.CheckClusterStatus(); err != nil {
			log.Printf("⚠️ Cluster status: %v", err)
		} else {
			fmt.Println("✅ Kubernetes cluster is healthy")
		}

		// Check MongoDB
		aggService, err := mongodb.NewAggregationService("mongodb://localhost:27017")
		if err != nil {
			log.Printf("⚠️ MongoDB connection: %v", err)
		} else {
			defer aggService.Close()
			fmt.Println("✅ MongoDB is accessible")

			// Count documents in each collection
			ctx := cmd.Context()
			if rawCount, err := aggService.DB.Collection("breweries_raw").CountDocuments(ctx, bson.M{}); err == nil {
				fmt.Printf("📊 Bronze layer (raw): %d documents\n", rawCount)
			}
			if cleanCount, err := aggService.DB.Collection("breweries_clean").CountDocuments(ctx, bson.M{}); err == nil {
				fmt.Printf("📊 Silver layer (clean): %d documents\n", cleanCount)
			}
			if aggCount, err := aggService.DB.Collection("breweries_aggregated").CountDocuments(ctx, bson.M{}); err == nil {
				fmt.Printf("📊 Gold layer (aggregated): %d documents\n", aggCount)
			}
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("brewctl v2.0.0")
		fmt.Println("Breweries Data Pipeline - Complete Implementation")
		fmt.Println("Built with Go, Airbyte, MongoDB, and Kubernetes")
	},
}

// ✅ ADICIONAR: Import necessário para bson.M

func init() {
	rootCmd.AddCommand(
		clusterInitCmd,
		deployConnectionsCmd,
		runAggregationsCmd,
		statusCmd,
		versionCmd,
	)
}

// Exemplo de uso completo
func main() {
	// 1. Criar cliente
	client := NewAirbyteClient("http://localhost:8000")

	// 2. Configurar todas as conexões
	if err := client.SetupConnections(); err != nil {
		fmt.Printf("❌ Setup failed: %v\n", err)
		return
	}

	// 3. (Opcional) Iniciar sincronização manual
	// connectionID := "sua-connection-id-aqui"
	// if err := client.SyncConnection(connectionID); err != nil {
	//     fmt.Printf("❌ Sync failed: %v\n", err)
	// }

	fmt.Println("✅ Airbyte pipeline configured successfully!")
}
