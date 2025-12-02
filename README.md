# Arbitrum-sequencer-decoder

A real-time arbitrage engine that detects cross-DEX price imbalances on Arbitrum using pre-consensus sequencer data. This system executes atomic arbitrage opportunities before they vanish, focusing on low-latency transaction processing and mathematical price impact estimation without EVM simulation.

## Project Status

This project is currently in the setup phase, implementing the documentation and test data structure as described in the architecture document. The actual Go implementation files (in `cmd/`, `pkg/` directories) will be created following the documented specifications.

## Architecture Overview

Based on the specification in `CLAUDE.md`, the system consists of:

- `cmd/` - Command-line applications for sequencer reading and transaction capture
- `pkg/decoder/` - DEX calldata decoders in pure Go
- `pkg/simulator/` - Price impact estimation without EVM
- `pkg/oracle/` - In-memory pool state tracking
- `pkg/arb-engine/` - Cross-pool arbitrage opportunity detection
- `pkg/executor/` - MEV-safe transaction bundle submission

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

## Test Data

The `testdata/sequencer/` directory contains realistic JSONL files with Arbitrum sequencer transaction data for:
- Uniswap V3 swaps
- Camelot V3 swaps
- Curve stableswap and cryptoswap transactions
- General Arbitrum sequencer feed data
- Mixed DEX transactions for integration testing

## Enterprise Development Workflow

This project follows enterprise-grade development practices:

- **[Git Workflow](GIT_WORKFLOW.md)** - Comprehensive branching strategy and commit standards
- **[Branch Protection](BRANCH_PROTECTION.md)** - Rules for protecting critical branches
- **[Code of Conduct](CODE_OF_CONDUCT.md)** - Standards for community interactions

## Getting Started

This project follows the principles from [HumanLayer.dev's Claude.md guide](https://www.humanlayer.dev/blog/writing-a-good-claude-md) for creating optimal coding environments for LLMs. The documentation and test data are designed to provide comprehensive context for implementing the actual Go code.

To begin implementation:
1. Familiarize yourself with the [Git Workflow](GIT_WORKFLOW.md) and development standards
2. Create the directory structure as outlined in `CLAUDE.md`
3. Implement the decoder packages following the specifications in `agent_docs/`
4. Use the test data in `testdata/sequencer/` for validation
5. Follow the build and testing guidelines in `agent_docs/build_and_test.md`
