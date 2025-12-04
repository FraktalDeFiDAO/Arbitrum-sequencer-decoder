# Development with Podman Compose

This project includes a complete development environment using Podman Compose. This setup allows you to run the Arbitrum sequencer decoder with all necessary supporting services in containers.

## Prerequisites

- Podman installed on your system
- Podman Compose installed (`pip install podman-compose`)
- For proper image resolution, ensure you have registry configuration. If you encounter short-name registry errors, create a registries configuration file:

```bash
mkdir -p ~/.config/containers
echo -e '[registries.search]\nregistries = ["docker.io", "quay.io"]' > ~/.config/containers/registries.conf
```

## Quick Start

The easiest way to start the development environment is using the provided development script:

```bash
# Start the development environment
./dev.sh dev
```

Or, if you prefer to use podman-compose directly:

```bash
# Build and start all services
podman-compose up --build

# Or start in detached mode
podman-compose up --build -d
```

## Services

The development environment includes these services:

- `sequencer-reader`: Main service that reads and processes Arbitrum sequencer transactions
- `sequencer-capture`: Service for capturing and storing transaction data
- `redis`: In-memory data store for caching and temporary data
- `postgres`: PostgreSQL database for persistent data storage
- `prometheus`: Metrics collection and monitoring
- `grafana`: Visualization and dashboard service
- `dev-tools`: Development container with Go environment

## Configuration

Configuration is controlled by the `.env` file in the project root. You can customize:

- `RPC_URL`: The Arbitrum RPC endpoint (defaults to mainnet)
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- PostgreSQL credentials
- Grafana admin password

## Useful Commands

The `dev.sh` script provides convenient commands for development:

```bash
# Start development environment
./dev.sh dev

# Start in detached mode
./dev.sh dev-detach

# View logs
./dev.sh logs

# Run tests
./dev.sh test

# Stop environment
./dev.sh down

# Open shell in development container
./dev.sh shell

# Connect to PostgreSQL
./dev.sh psql
```

## Development Workflow

For active development, use this workflow:

1. Run the dev tools container to have a Go development environment:
   ```bash
   ./dev.sh shell
   ```

2. The source code is mounted in the dev-tools container, so changes made on the host are immediately available inside the container.

3. You can build and run the application directly in the dev container:
   ```bash
   go build -o sequencer-reader ./cmd/sequencer-reader
   ./sequencer-reader -rpc $RPC_URL
   ```

4. Install and use development tools inside the container:
   ```bash
   # Install linters and security tools
   podman-compose exec dev-tools sh -c "apk add --no-cache golangci-lint && go install github.com/securego/gosec/v2/cmd/gosec@latest"

   # Run code quality checks
   podman-compose exec dev-tools sh -c "cd /app && golangci-lint run"

   # Run security scans
   podman-compose exec dev-tools sh -c "cd /app && gosec ./..."
   ```

5. Access services:
   - Application metrics: http://localhost:8080/metrics
   - Grafana dashboard: http://localhost:3000 (admin/admin by default)
   - Prometheus: http://localhost:9090
   - PostgreSQL: localhost:5432

## Monitoring

Grafana comes preconfigured with a dashboard for the sequencer metrics. After starting the environment:

1. Visit http://localhost:3000
2. Log in with admin/admin (or your custom password)
3. The Arbitrum Sequencer Decoder dashboard will show transaction processing metrics

## Environment Variables

You can override default settings by creating or modifying the `.env` file in the project root.

## Podman-in-Podman Development Setup

This project is designed to work with a containerized development environment where development tools run inside the `dev-tools` container. This approach provides:

- Consistent development environment across all team members
- Properly configured tooling with all dependencies
- Isolation of development tools from your host system
- Pre-configured networking and service dependencies

### Key Features of the Containerized Development Environment

1. The `dev-tools` container includes a complete Go development environment
2. Source code is mounted at `/app` inside the container
3. All services (redis, postgres, etc.) are available via the internal network
4. You can run linters, security tools, builds, and tests from within the container

### Running Development Commands in the Container

Instead of running development tools directly on your host system, use the following approach:

```bash
# Run linters inside the dev-tools container
podman-compose exec dev-tools sh -c "cd /app && golangci-lint run"

# Run security scans
podman-compose exec dev-tools sh -c "cd /app && gosec ./..."

# Build the application
podman-compose exec dev-tools sh -c "cd /app && go build -o sequencer-reader ./cmd/sequencer-reader"

# Run tests
podman-compose exec dev-tools sh -c "cd /app && go test ./..."

# Interactive development
podman-compose exec dev-tools sh
```

### Benefits of Using Containerized Development

- **Consistency**: All developers work in the same environment
- **Isolation**: No pollution of your host system with development dependencies
- **Reproducibility**: Build and test results are consistent across environments
- **Complete Setup**: All services are already configured and running

## Troubleshooting

If you encounter issues:

1. Make sure all services are running:
   ```bash
   podman-compose ps
   ```

2. Check logs for specific services:
   ```bash
   podman-compose logs sequencer-reader
   podman-compose logs postgres
   ```

3. If containers fail to start, try resetting the environment:
   ```bash
   ./dev.sh reset
   ```

4. If you get container name conflicts like "arb-* is not a valid container", this typically means old containers exist with the same names. To resolve this:
   ```bash
   # Stop and remove all containers
   ./dev.sh reset

   # Or manually remove containers
   podman rm -f sequencer-reader sequencer-capture arb-redis arb-postgres arb-prometheus arb-grafana arb-dev-tools 2>/dev/null || true

   # Then restart the environment
   ./dev.sh dev
   ```