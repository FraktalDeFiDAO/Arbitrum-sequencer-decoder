# Arbitrum-sequencer-decoder

A real-time arbitrage engine that detects cross-DEX price imbalances on Arbitrum using pre-consensus sequencer data. This system executes atomic arbitrage opportunities before they vanish, focusing on low-latency transaction processing and mathematical price impact estimation without EVM simulation.

## Project Status

This project is currently in the implementation phase, following the documented project plan in [PROJECT_PLAN.md](docs/project_planning/PROJECT_PLAN.md). The foundational architecture has been established with Go module, directory structure, CLI framework, and common types implemented as detailed in the [PROJECT_PROGRESS.md](docs/project_planning/PROJECT_PROGRESS.md).

Current implementation includes:
- Go module initialized with required dependencies
- Directory structure created: `cmd/sequencer-reader`, `cmd/sequencer-capture`, `pkg/decoder`, `pkg/simulator`, `pkg/oracle`, `pkg/arb-engine`, `pkg/executor`, `pkg/types`
- CLI framework with configuration and logging in `cmd/sequencer-reader/main.go`
- Transaction capture functionality in `cmd/sequencer-capture/main.go`
- Common types, interfaces, and error definitions in `pkg/types/`
- Ready for Phase 2: Core Infrastructure implementation

## Architecture Overview

Based on the specification in `CLAUDE.md`, the system consists of:

- `cmd/` - Command-line applications for sequencer reading and transaction capture
- `pkg/decoder/` - DEX calldata decoders in pure Go
- `pkg/simulator/` - Price impact estimation without EVM
- `pkg/oracle/` - In-memory pool state tracking
- `pkg/arb-engine/` - Cross-pool arbitrage opportunity detection
- `pkg/executor/` - MEV-safe transaction bundle submission
- `agents/auditing/` - Comprehensive auditing system for monitoring and validation

## Supported DEXes

- Uniswap V2/V3/V4
- Camelot V2/V3
- Ramses
- Curve (Stableswap/Cryptoswap)
- Balancer V2/V3
- Kyber Classic/Elastic

## Documentation Structure

The `agent_docs/` directory contains detailed documentation for each component:
- DEX-specific calldata decoding instructions
- Build and testing procedures
- Sequencer feed replay mechanisms
- Price simulation algorithms

## Auditing System

The `agents/auditing/` directory contains a comprehensive auditing system that includes:
- Decoder Auditor - Monitors decoder performance and accuracy
- Arbitrage Auditor - Validates arbitrage opportunities for profitability and risk
- System Health Auditor - Monitors overall system health and component status
- Audit Manager - Coordinates all auditing activities

## Test Data

The `testdata/sequencer/` directory contains realistic JSONL files with Arbitrum sequencer transaction data for:
- Uniswap V3 swaps
- Camelot V3 swaps
- Curve stableswap and cryptoswap transactions
- General Arbitrum sequencer feed data
- Mixed DEX transactions for integration testing

## Enterprise Development Workflow

This project follows enterprise-grade development practices:

- **[Git Workflow](docs/general/GIT_WORKFLOW.md)** - Comprehensive branching strategy and commit standards
- **[Branch Protection](docs/general/BRANCH_PROTECTION.md)** - Rules for protecting critical branches
- **[Code of Conduct](CODE_OF_CONDUCT.md)** - Standards for community interactions

## Getting Started

This project follows the principles from [HumanLayer.dev's Claude.md guide](https://www.humanlayer.dev/blog/writing-a-good-claude-md) for creating optimal coding environments for LLMs. The documentation and test data are designed to provide comprehensive context for implementing the actual Go code.

### Prerequisites

- Podman installed on your system
- Podman Compose installed (`pip install podman-compose`)
- For proper image resolution, ensure you have registry configuration:

```bash
mkdir -p ~/.config/containers
echo -e '[registries.search]\nregistries = ["docker.io", "quay.io"]' > ~/.config/containers/registries.conf
```

### Development Environment (Recommended)

The project includes a complete containerized development environment. To start:

```bash
# Start the development environment with all services
./dev.sh dev

# Or run in detached mode
./dev.sh dev-detach

# Access the development tools container
./dev.sh shell

# View logs
./dev.sh logs
```

This will start all necessary services (sequencer-reader, sequencer-capture, redis, postgres, prometheus, grafana, and dev-tools) with proper networking and volumes.

### Development Workflow

1. Start the development environment:
   ```bash
   ./dev.sh dev
   ```

2. Access the dev-tools container for development:
   ```bash
   ./dev.sh shell
   ```

3. Run code quality checks inside the container:
   ```bash
   # Run linter
   podman-compose exec dev-tools sh -c "cd /app && golangci-lint run"

   # Run security scanner
   podman-compose exec dev-tools sh -c "cd /app && gosec ./..."
   ```

4. Build and test your changes:
   ```bash
   # Inside the dev container
   cd /app
   go build -o sequencer-reader ./cmd/sequencer-reader
   go test ./...
   ```

To continue implementation:
1. Familiarize yourself with the [Git Workflow](docs/general/GIT_WORKFLOW.md) and development standards
2. Review the [Project Plan](docs/project_planning/PROJECT_PLAN.md) for the detailed roadmap
3. Check the [Project Progress](docs/project_planning/PROJECT_PROGRESS.md) to see what's been completed
4. Implement the decoder packages following the specifications in `agent_docs/`
5. Use the test data in `testdata/sequencer/` for validation
6. Follow the build and testing guidelines in `agent_docs/build_and_test.md`

## Current Implementation Status

Based on the [PROJECT_PROGRESS.md](docs/project_planning/PROJECT_PROGRESS.md):
- ✅ Phase 1.1: Go module and directory structure
- ✅ Phase 1.2: CLI framework in cmd/sequencer-reader
- ✅ Phase 1.3: Common types and interfaces in pkg/types
- ✅ Phase 1.4: Transaction capture in cmd/sequencer-capture
- Next: Phase 2: Core Infrastructure (Tasks 2.1-2.4)
- Then: Phase 3: DEX Decoder Implementation (Tasks 3.1-3.28)
