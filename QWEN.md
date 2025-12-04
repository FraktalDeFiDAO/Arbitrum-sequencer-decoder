# Arbitrum-sequencer-decoder

## Project Overview

Arbitrum-sequencer-decoder is a real-time arbitrage engine that detects cross-DEX (Decentralized Exchange) price imbalances on the Arbitrum One blockchain using pre-consensus sequencer data. The system executes atomic arbitrage opportunities before they vanish, with latency being critical for success.

The project is built in Go (version 1.25+) and focuses on reading raw sequencer transaction feeds from Arbitrum to decode DEX calldata, simulate trades without an EVM, track pool states in memory, find arbitrage opportunities across similar pools, and submit MEV-safe transaction bundles.

## Architecture

The core components of the system are organized as follows:
- `cmd/sequencer-reader/` – Reads sequencer stream
- `cmd/sequencer-capture/` – Captures and stores transactions for analysis
- `pkg/decoder/` – Decodes DEX calldata in pure Go
- `pkg/simulator/` – Estimates post-swap prices without EVM
- `pkg/oracle/` – Tracks in-memory pool states
- `pkg/arb-engine/` – Finds arbitrage across like pools
- `pkg/executor/` – Submits MEV-safe bundles
- `pkg/types/` – Common types and interfaces

## Supported DEXes

The system supports multiple decentralized exchanges:
- Uniswap V2/V3/V4
- Camelot V2/V3
- Ramses
- Curve
- Balancer V2/V3
- Kyber Classic/Elastic

## Core Features

### Sequencer Data Processing
- Uses raw, unsigned, pre-consensus Ethereum transactions sent to the `SequencerInbox` contract
- Works with data before execution when logs and events are unavailable
- Only calldata and recipient address information are available
- Accounts for the fact that 10-30% of transactions may be dropped before inclusion

### Calldata Decoding
- Directly parses calldata for each supported DEX without relying on logs/events
- Supports complex path routing for multi-hop swaps
- Implements stateless mathematical models for price impact estimation

### Price Simulation
- Uses pure Go mathematical implementations (e.g., Uniswap tick math, Curve invariants)
- No external runtimes (no Python, JS, or EVM forks) required
- Efficient price impact calculations using analytical formulas

### Data Replay
- Includes functionality to replay historical or live Arbitrum sequencer data
- Allows testing of decoding, simulation, and arbitrage logic without waiting for real-time opportunities
- Essential for development, regression testing, and performance profiling

## Building and Running

### Prerequisites
- Go version 1.25+
- Dependencies: Only `go-ethereum` (lite), stdlib, and embedded ABIs

### Build Commands
```bash
# Build sequencer reader
go build -o bin/sequencer-reader ./cmd/sequencer-reader

# Build sequencer capture
go build -o bin/sequencer-capture ./cmd/sequencer-capture
```

### Running
```bash
# Run sequencer reader
./bin/sequencer-reader -rpc https://arb1.arbitrum.io/rpc

# Run sequencer capture
./bin/sequencer-capture -rpc https://arb1.arbitrum.io/rpc -output ./testdata/sequencer/captured_txs.jsonl
```

## Development Status

Based on the [PROJECT_PROGRESS.md](docs/project_planning/PROJECT_PROGRESS.md):
- ✅ Phase 1.1: Go module and directory structure
- ✅ Phase 1.2: CLI framework in cmd/sequencer-reader
- ✅ Phase 1.3: Common types and interfaces in pkg/types
- ✅ Phase 1.4: Transaction capture in cmd/sequencer-capture
- Next: Phase 2: Core Infrastructure (Tasks 2.1-2.4)
- Then: Phase 3: DEX Decoder Implementation (Tasks 3.1-3.28)

## Development Conventions

### Testing
- The project includes comprehensive test files in the `testdata/sequencer` directory
- Supports replay testing with various DEX-specific formats
- All functionality should be validated with sequencer samples

### Documentation
- Detailed documentation for each DEX integration is available in the `agent_docs/` directory
- DEX-specific calldata formats and decoding instructions are documented
- Build and test procedures are standardized

## Key Files and Directories

- `docs/project_planning/PROJECT_PLAN.md` - Detailed project roadmap and task breakdown
- `docs/project_planning/PROJECT_PROGRESS.md` - Current implementation status and task tracking
- `CLAUDE.md` - Main project overview and architecture description
- `agent_docs/` - Detailed documentation for each supported DEX and development workflows
- `testdata/sequencer/` - Sample data for various DEX swaps including Camelot V3, Curve, and other transaction types
- `cmd/` - Command-line applications for reading and capturing sequencer data
- `pkg/` - Core packages for decoding, simulation, oracle, arbitrage engine, and execution

## Data Sources

The system works with Arbitrum sequencer feeds from endpoints like `https://arb1.arbitrum.io/rpc` using `eth_call`-like streaming via `eth_getBlockByNumber("pending", false)` or dedicated sequencer proxies.

## Testing and Validation

The project includes replay functionality that allows historical or live sequencer data to be used for testing decoding, simulation, and arbitrage logic without waiting for real-time opportunities. This is essential for development, regression testing, and performance profiling.

The test data includes examples for:
- Camelot V3 swaps
- Curve swaps
- Other DEX transaction types

The system is designed to be reproducible, fast, and compatible with CI environments.