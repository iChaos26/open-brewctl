# Copilot Instructions for brewctl
# !TODO: PUT IMAGES OF CLUSTER DEPLOY
## Project Overview
- **brewctl** is a CLI tool for managing deployments and clusters, with a focus on Airbyte, MongoDB, and monitoring stacks (Grafana, Prometheus).
- The codebase is organized by major components: `cmd/` (CLI commands), `internal/` (core logic), `deplyoments/` (configuration files), and `pkg/` (shared types/utilities).

## Architecture & Key Patterns
- **Command Structure:**
  - Each CLI command is in its own subdirectory under `cmd/` (e.g., `cmd/airbyte/`, `cmd/cluster/`).
  - The entrypoint is `cmd/root.go`, which wires up subcommands.
- **Internal Logic:**
  - Business logic and integrations are in `internal/`, grouped by domain (e.g., `internal/airbyte/client.go`, `internal/kube/helm.go`).
  - Use Go interfaces and struct composition for extensibility.
- **Configuration:**
  - Deployment configs (Helm values, Kind configs) are in `deplyoments/`.
  - Reference these files for cluster setup and service deployment.
- **Utilities & Types:**
  - Shared helpers and type definitions are in `pkg/types/` and `pkg/utils/`.

## Developer Workflows
- **Build:**
  - Standard Go build: `go build ./...` from the project root.
- **Test:**
  - Run all tests: `go test ./...`
  - No custom test runners detected; use Go's built-in tooling.
- **Debug:**
  - Main entrypoint for debugging is `cmd/root.go`.
  - For CLI command debugging, run with verbose flags if implemented.

## Conventions & Patterns
- **File Naming:**
  - Command files use `*.go` and are grouped by feature.
  - Config files use `*-values.yaml` or `*-config.yaml`.
- **Error Handling:**
  - Use Go error returns; propagate errors up to CLI for user feedback.
- **External Integrations:**
  - Integrates with Kubernetes (Kind, Helm), Airbyte, MongoDB, Grafana, Prometheus.
  - All integration logic is in `internal/`.

## Examples
- To add a new CLI command: create a new subdirectory in `cmd/`, implement the command, and register it in `cmd/root.go`.
- To add a new deployment config: place the YAML file in `deplyoments/` and reference it in the relevant internal logic.

## Key Files & Directories
- `cmd/root.go`: CLI entrypoint and command registration
- `internal/airbyte/`: Airbyte integration logic
- `internal/kube/`: Kubernetes/Helm/Kind logic
- `deplyoments/`: Deployment configuration files
- `pkg/types/`, `pkg/utils/`: Shared types and utilities

# Project Structure V1
# Brewctl - Breweries Data Pipeline

A CLI tool to manage Airbyte, MongoDB and monitoring stack for breweries data pipeline.

## Features

- Deploy a local Kubernetes cluster (Kind)
- Deploy Airbyte for data ingestion
- Deploy MongoDB for data storage
- Deploy Prometheus and Grafana for monitoring
- Setup Airbyte connections for Brewery API
- Run data aggregations and transformations

## Prerequisites

- Docker
- Kind
- Helm
- Go 1.16+

## Usage

1. Build the CLI:
   ```bash
   go build -o brewctl cmd/root.go
---
brewctl/
├── cmd/
│   └── brewctl/
│       └── main.go
├── internal/
│   ├── kube/
│   │   ├── kind.go
│   │   └── helm.go
│   ├── airbyte/
│   │   ├── client.go
│   │   ├── deploy.go
│   │   └── connections.go
│   ├── monitoring/
│   │   ├── prometheus.go
│   │   └── grafana.go
│   └── mongodb/
│       ├── client.go
│       └── aggregations.go
├── scripts/
│   ├── setup-and-run.sh
│   └── mongodb-aggregations.js
├── deployments/
│   ├── kind-config.yaml
│   ├── airbyte-values.yaml
│   ├── mongodb-values.yaml
│   └── monitoring-values.yaml
├── pkg/
│   └── types/
│       └── types.go
├── go.mod
├── go.sum
└── README.md

---
