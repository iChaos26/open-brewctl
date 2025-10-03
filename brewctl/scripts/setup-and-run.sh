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

# Build the CLI
echo "🔨 Building brewctl CLI..."
go mod tidy
go build -o brewctl ./cmd/brewctl

# Initialize cluster
echo "🚀 Initializing cluster..."
./brewctl cluster-init

# Wait for services to be ready with better checking
echo "⏳ Waiting for services to be ready..."

echo "📋 Waiting for Kubernetes nodes..."
kubectl wait --for=condition=Ready nodes --all --timeout=180s

echo "📋 Waiting for Airbyte pods (this can take 3-5 minutes)..."
kubectl wait --for=condition=Ready pod -l app.kubernetes.io/name=airbyte --timeout=300s

echo "📋 Waiting for MongoDB pod..."
kubectl wait --for=condition=Ready pod -l app.kubernetes.io/name=mongodb --timeout=180s

echo "📋 Waiting for monitoring pods..."
kubectl wait --for=condition=Ready pod -l app=grafana --timeout=180s
kubectl wait --for=condition=Ready pod -l app=prometheus-server --timeout=180s

# Additional buffer for services to stabilize
echo "💤 Allowing additional time for services to stabilize..."
sleep 60

echo ""
echo "🔍 Checking final status..."
kubectl get pods --all-namespaces

# Deploy connections
echo "🔗 Deploying Airbyte connections..."
if ! ./brewctl deploy-connections; then
    echo "❌ Failed to deploy Airbyte connections."
    echo "💡 You can try running './brewctl deploy-connections' manually later."
fi

# Run aggregations
echo "🔄 Running data aggregations..."
if ! ./brewctl run-aggregations; then
    echo "❌ Data aggregations failed."
    echo "💡 This might be because Airbyte sync hasn't completed yet."
    echo "💡 You can try running './brewctl run-aggregations' manually later."
fi

echo ""
echo "✅ Setup completed (services may still be starting)..."
echo ""
echo "🌐 Access URLs (wait 2-3 minutes after setup):"
echo "   Airbyte:     http://localhost:8000"
echo "   Grafana:     http://localhost:3000 (admin/admin)"
echo "   Prometheus:  http://localhost:9090"
echo "   MongoDB:     localhost:27017"
echo ""
echo "🔧 If URLs don't work, run: ./scripts/check-health.sh"