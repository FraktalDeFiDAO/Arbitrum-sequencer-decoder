# Sequencer Feed Replay

## Purpose
Replay historical or live Arbitrum sequencer data to test decoding, simulation, and arbitrage logic **without waiting for real-time opportunities**. This is essential for development, regression testing, and performance profiling.

## Data Source
The Arbitrum sequencer feed is a stream of **raw, unsigned, pre-consensus Ethereum transactions** sent to the `SequencerInbox` contract.
- **Endpoint**: `https://arb1.arbitrum.io/rpc` (use `eth_call`-like streaming via `eth_getBlockByNumber("pending", false)` or dedicated sequencer proxies)
- **Format**: Standard Ethereum JSON-RPC transaction objects (`to`, `data`, `value`, `gas`, etc.)
- **Non-final**: ~10–30% of transactions may be dropped before inclusion

## Local Capture
Capture live sequencer traffic:
```bash
go run ./cmd/sequencer-capture --output ./testdata/sequencer/live_20251202.jsonl
```

## Replay Commands
```bash
# Replay specific DEX transactions
go run ./cmd/replay --input ./testdata/sequencer/camelot_v3_swaps.jsonl --decoder camelot_v3

# Replay all transactions in file
go run ./cmd/replay --input ./testdata/sequencer/uniswap_swaps.jsonl

# Performance test with replay
go run ./cmd/perf-test --input ./testdata/sequencer/large_batch.jsonl --concurrency 10
```

## Data Format (JSONL)
Each line contains a JSON object with transaction data:
```json
{
  "hash": "0x...",
  "from": "0x...",
  "to": "0x...",
  "data": "0x...",
  "value": "0x...",
  "gas": "0x...",
  "timestamp": "2025-12-02T10:00:00Z",
  "blockNumber": "0x..."
}
```

## Performance Considerations
- **Batch replay**: Process multiple transactions together for efficiency
- **Parallel decoding**: Run multiple decoder goroutines to match live speeds
- **Memory management**: Stream processing to avoid loading entire files in memory
- **Cache warming**: Pre-load pool states to reduce initial replay latency