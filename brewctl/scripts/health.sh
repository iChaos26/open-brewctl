#!/bin/bash

echo "🏥 Cluster Health Check"
echo "======================="

echo ""
echo "📋 Nodes:"
kubectl get nodes

echo ""
echo "🐳 Pods Status:"
kubectl get pods --all-namespaces

echo ""
echo "🔗 Services:"
kubectl get services

echo ""
echo "📦 Airbyte Pods:"
kubectl get pods -l app.kubernetes.io/name=airbyte

echo ""
echo "🍃 MongoDB Pods:"
kubectl get pods -l app.kubernetes.io/name=mongodb

echo ""
echo "📊 Monitoring Pods:"
kubectl get pods -l app=grafana
kubectl get pods -l app=prometheus

echo ""
echo "⏳ Waiting for pods to be ready..."
kubectl wait --for=condition=Ready pods --all --timeout=300s

echo ""
echo "✅ Health check completed"