# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Arbitrum-sequencer-decoder is a real-time arbitrage engine that detects cross-DEX price imbalances on Arbitrum One using pre-consensus sequencer data. Built in Go 1.25+, it decodes DEX calldata, simulates trades without EVM, and submits MEV-safe bundles. Latency is critical.

Key constraint: Works with raw, unsigned transactions before execution—only calldata and recipient address are available (no logs/events). 10-30% of transactions may be dropped before inclusion.

## Build and Test Commands

```bash
# Build all binaries
go build -tags=arbitrum -o bin/arb-engine ./cmd/...

# Run all tests
go test ./... -v

# Run specific decoder test
go test ./pkg/decoder/uniswap_v3 -run TestUniswapV3Decoder -v

# Run benchmarks
go test ./pkg/simulator -bench=. -benchmem

# Format and lint before committing
go fmt ./...
golangci-lint run
```

## Architecture

```
cmd/
├── sequencer-reader/     # Reads live sequencer stream
├── sequencer-capture/    # Captures transactions to JSONL for testing
├── capture-dex-transactions/
├── query-dex-transactions/
└── query-historical-dex-transactions/

pkg/
├── types/               # Core types: Transaction, DecodedAction, Pool, PoolState
├── classifier/          # Identifies DEX txs by address + function signature
├── decoder/             # DEX-specific calldata decoders
│   ├── interface.go     # Decoder interface all DEXes implement
│   ├── uniswap_v3/      # Uniswap V3 decoder (implemented)
│   └── camelot_v3/      # Camelot V3 decoder (implemented)
├── simulator/           # Price impact estimation (pure Go math, no EVM)
│   ├── uniswap_v3/      # Tick math implementation
│   └── camelot_v3/
├── oracle/              # In-memory pool state tracking
├── blockchain/          # RPC client and query utilities
├── arb-engine/          # Cross-pool arbitrage detection
└── executor/            # MEV-safe bundle submission
```

## Core Interfaces

**Decoder interface** (`pkg/decoder/interface.go`):
```go
type Decoder interface {
    Matches(tx *ethtypes.Transaction, toAddress string) bool
    Decode(tx *ethtypes.Transaction, toAddress string) ([]types.DecodedAction, error)
    Protocol() types.ProtocolType
}
```

Each DEX decoder implements this interface. Use `decoder.MatchesSignature()` helper to check function signatures.

**Key types** (`pkg/types/types.go`):
- `Transaction` - Decoded sequencer transaction with `Actions []DecodedAction`
- `DecodedAction` - Single DEX operation (swap, add/remove liquidity)
- `Pool`, `PoolState` - Pool metadata and current reserves/liquidity
- `PoolType`, `ProtocolType` - Enums for supported DEXes

## Supported DEXes

Implemented: Uniswap V3, Camelot V3
Planned: Uniswap V2/V4, Curve, Balancer V2/V3, Ramses, Kyber Classic/Elastic

Known router addresses are in `pkg/classifier/classifier.go`.

## Adding a New DEX Decoder

1. Create `pkg/decoder/<dex_name>/` directory
2. Implement `abi_bindings.go` with router addresses and function signatures
3. Implement `decoder.go` satisfying the `Decoder` interface
4. Add price simulation math in `pkg/simulator/<dex_name>/`
5. Add router address to `pkg/classifier/classifier.go`
6. Create tests using real transaction data from `testdata/sequencer/`

Reference `agent_docs/<dex>_calldata.md` for DEX-specific encoding details.

## Test Data

`testdata/sequencer/` contains JSONL files with real Arbitrum transactions:
- `uniswap_v3_*.jsonl` - Uniswap V3 swaps
- `camelot_v3_*.jsonl` - Camelot V3 swaps
- `*_factory_transactions.jsonl`, `*_router_transactions.jsonl`

Use `cmd/sequencer-capture` to capture new test data:
```bash
./bin/sequencer-capture -rpc https://arb1.arbitrum.io/rpc -output ./testdata/sequencer/captured_txs.jsonl
```

## Development Status

See `docs/project_planning/PROJECT_PROGRESS.md` for detailed task tracking. Current state:
- Phase 1 (Setup): Complete
- Phase 2 (Core Infrastructure): Complete
- Phase 3 (DEX Decoders): Uniswap V3 and Camelot V3 complete; Curve, Balancer, Ramses, Kyber pending