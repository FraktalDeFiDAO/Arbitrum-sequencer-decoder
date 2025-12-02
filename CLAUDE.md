# Arbitrum Real-Time Arbitrage Engine

## WHY
Detect cross-DEX price imbalances on Arbitrum using pre-consensus sequencer data and execute atomic arbitrage before opportunities vanish. Latency matters—every millisecond counts.

## WHAT
- **Language**: Go 1.25+
- **Chain**: Arbitrum One
- **Input**: Raw sequencer transaction feed (non-final, may be dropped)
- **Supported DEXes**: Uniswap V2/V3/V4, Camelot V2/V3, Ramses, Curve, Balancer V2/V3, Kyber Classic/Elastic
- **Core Structure**:
  - `cmd/sequencer-reader/` – reads sequencer stream
  - `pkg/decoder/` – decodes DEX calldata in pure Go
  - `pkg/simulator/` – estimates post-swap prices without EVM
  - `pkg/oracle/` – tracks in-memory pool states
  - `pkg/arb-engine/` – finds arbitrage across like pools
  - `pkg/executor/` – submits MEV-safe bundles

## HOW
- The system uses **no external runtimes** (no Python, JS, or EVM forks).
- All price impact logic is **pure Go math** (e.g., Uniswap tick math, Curve invariants).
- Calldata is parsed **directly**—logs/events are not available pre-execution.
- See `agent_docs/` for DEX-specific formats, build instructions, and testing guides.

## Project Status
This project is in the setup phase, following the principles from [HumanLayer.dev's Claude.md guide](https://www.humanlayer.dev/blog/writing-a-good-claude-md) for creating optimal coding environments for LLMs. The documentation and test data have been created per the specifications, and the actual Go implementation will follow the documented architecture.

## Completed Components
- **agent_docs/**: Complete documentation for all supported DEXes, build and test procedures, and sequencer replay mechanisms
- **testdata/sequencer/**: Realistic JSONL files with Arbitrum sequencer transaction data for various DEXes
- **QWEN.md**: Comprehensive project overview for LLM context
- **README.md**: Project status and getting started information

