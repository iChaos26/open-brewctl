#!/bin/bash

echo "🔍 Checking port usage..."
echo "Port 80:"
sudo lsof -i :80 || echo "✅ Port 80 is free or not accessible"

echo ""
echo "Port 443:"
sudo lsof -i :443 || echo "✅ Port 443 is free or not accessible"

echo ""
echo "Required ports for brewctl:"
echo "8000 (Airbyte):"
sudo lsof -i :8000 || echo "✅ Free"
echo "3000 (Grafana):"
sudo lsof -i :3000 || echo "✅ Free"
echo "9090 (Prometheus):"
sudo lsof -i :9090 || echo "✅ Free"
echo "27017 (MongoDB):"
sudo lsof -i :27017 || echo "✅ Free"