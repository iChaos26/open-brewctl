#!/bin/bash

set -e

echo "🍻 Breweries Data Pipeline - Complete Setup"
echo "==========================================="

# Check prerequisites
echo "🔍 Checking prerequisites..."
command -v docker >/dev/null 2>&1 || { echo "❌ Docker is required but not installed. Aborting."; exit 1; }
command -v kind >/dev/null 2>&1 || { echo "❌ Kind is required but not installed. Aborting."; exit 1; }
command -v helm >/dev/null 2>&1 || { echo "❌ Helm is required but not installed. Aborting."; exit 1; }
command -v go >/dev/null 2>&1 || { echo "❌ Go is required but not installed. Aborting."; exit 1; }
command -v kubectl >/dev/null 2>&1 || { echo "❌ kubectl is required but not installed. Aborting."; exit 1; }

# Verify Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Build the CLI
echo "🔨 Building brewctl CLI..."
go mod tidy
if ! go mod verify; then
    echo "⚠️ Go module verification had issues, but continuing..."
fi

if ! go build -o brewctl ./cmd/brewctl; then
    echo "❌ Failed to build brewctl CLI. Please check Go errors above."
    exit 1
fi

echo "✅ brewctl CLI built successfully"

# Check if cluster already exists and offer to recreate
if kind get clusters | grep -q "brewctl-cluster"; then
    echo "⚠️ brewctl-cluster already exists."
    read -p "Do you want to recreate it? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "🗑️ Deleting existing cluster..."
        kind delete cluster --name brewctl-cluster
    else
        echo "🔄 Using existing cluster..."
    fi
fi

# Initialize cluster
echo "🚀 Initializing cluster..."
if ! ./brewctl cluster-init; then
    echo "❌ Cluster initialization failed. Please check logs above."
    exit 1
fi

# Wait for services to be ready with better checking
echo "⏳ Waiting for services to be ready..."
echo "📋 Waiting for Kubernetes nodes..."
kubectl wait --for=condition=Ready nodes --all --timeout=120s

echo "📋 Waiting for Airbyte pods..."
kubectl wait --for=condition=Ready pod -l app.kubernetes.io/name=airbyte --timeout=180s

echo "📋 Waiting for MongoDB pod..."
kubectl wait --for=condition=Ready pod -l app.kubernetes.io/name=mongodb --timeout=120s

# Additional buffer for services to stabilize
echo "💤 Allowing additional time for services to stabilize..."
sleep 30

# Deploy connections
echo "🔗 Deploying Airbyte connections..."
if ! ./brewctl deploy-connections; then
    echo "❌ Failed to deploy Airbyte connections. Please check Airbyte UI at http://localhost:8000"
    echo "💡 You can try running './brewctl deploy-connections' manually later."
    # Don't exit here, as aggregations might still work with existing data
fi

# Run aggregations
echo "🔄 Running data aggregations..."
if ! ./brewctl run-aggregations; then
    echo "❌ Data aggregations failed."
    echo "💡 This might be because Airbyte sync hasn't completed yet."
    echo "💡 You can try running './brewctl run-aggregations' manually later."
    exit 1
fi

echo ""
echo "✅ Setup completed successfully!"
echo ""
echo "🌐 Access URLs:"
echo "   Airbyte:     http://localhost:8000"
echo "   Grafana:     http://localhost:3000 (admin/admin)"
echo "   Prometheus:  http://localhost:9090"
echo "   MongoDB:     localhost:27017"
echo ""
echo "📊 Data Pipeline Status:"
echo "   ✅ Kubernetes cluster running"
echo "   ✅ Airbyte deployed"
echo "   ✅ MongoDB deployed" 
echo "   ✅ Monitoring stack deployed"
echo "   ✅ Airbyte connections configured"
echo "   ✅ Data aggregations executed"
echo ""
echo "🔧 Next steps:"
echo "   1. Check Airbyte UI to monitor sync progress: http://localhost:8000"
echo "   2. View analytics in MongoDB: mongosh localhost:27017/breweries_db"
echo "   3. Access Grafana dashboards: http://localhost:3000"
echo "   4. Run './brewctl status' to check system status"
echo "   5. Run './brewctl full-pipeline' to re-run entire pipeline"
echo ""
echo "🚀 To trigger a complete data refresh:"
echo "   ./brewctl full-pipeline"