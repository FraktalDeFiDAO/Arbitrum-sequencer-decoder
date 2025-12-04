#!/bin/bash

# Arbitrum Sequencer Decoder Development Utilities

case "$1" in
  "build")
    echo "Building the Arbitrum sequencer decoder binaries..."
    go build -o bin/sequencer-reader ./cmd/sequencer-reader
    go build -o bin/sequencer-capture ./cmd/sequencer-capture
    ;;
  "test")
    echo "Running tests..."
    # Use local cache to avoid permission issues in sandboxed environments
    export GOCACHE=${GOCACHE:-$(pwd)/.gocache}
    mkdir -p "$GOCACHE"
    go test ./...
    ;;
  "lint")
    echo "Running linter inside dev-tools container..."
    podman-compose exec dev-tools sh -c "cd /app && golangci-lint run"
    ;;
  "security")
    echo "Running security scan inside dev-tools container..."
    podman-compose exec dev-tools sh -c "cd /app && gosec ./..."
    ;;
  "dev")
    echo "Starting development environment with Podman Compose..."
    podman-compose up --build
    ;;
  "dev-detach")
    echo "Starting development environment in detached mode..."
    podman-compose up --build -d
    ;;
  "logs")
    echo "Showing logs..."
    podman-compose logs -f
    ;;
  "down")
    echo "Stopping development environment..."
    podman-compose down
    ;;
  "reset")
    echo "Stopping and removing all containers and volumes..."
    podman-compose down -v
    ;;
  "shell")
    echo "Opening shell in the dev-tools container..."
    podman-compose exec dev-tools sh
    ;;
  "psql")
    echo "Connecting to PostgreSQL database..."
    podman-compose exec postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"
    ;;
  *)
    echo "Usage: $0 {build|test|lint|security|dev|dev-detach|logs|down|reset|shell|psql}"
    echo ""
    echo "Commands:"
    echo "  build       - Build the Go binaries"
    echo "  test        - Run all tests"
    echo "  lint        - Run linter inside dev-tools container"
    echo "  security    - Run security scan inside dev-tools container"
    echo "  dev         - Start development environment"
    echo "  dev-detach  - Start development environment in background"
    echo "  logs        - Show container logs"
    echo "  down        - Stop development environment"
    echo "  reset       - Stop and remove all containers and volumes"
    echo "  shell       - Open shell in dev-tools container"
    echo "  psql        - Connect to PostgreSQL database"
    exit 1
    ;;
esac
