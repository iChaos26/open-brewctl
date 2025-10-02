#!/bin/bash

set -e

echo "🍻 Quick Start - Breweries Pipeline"
echo "==================================="

# Build
echo "🔨 Building CLI..."
go mod tidy
go build -o brewctl ./cmd/brewctl

# Test basic functionality
echo "🚀 Testing CLI..."
./brewctl version

echo ""
echo "✅ Ready to use! Available commands:"
echo "   ./brewctl cluster-init      # Start Kubernetes cluster"
echo "   ./brewctl full-pipeline     # Run complete pipeline"
echo "   ./brewctl deploy-connections # Setup Airbyte connections"
echo "   ./brewctl run-aggregations  # Run MongoDB aggregations"
echo "   ./brewctl status           # Check system status"