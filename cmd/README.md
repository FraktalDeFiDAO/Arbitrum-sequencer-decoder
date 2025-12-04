# Arbitrum Sequencer Decoder Commands

This directory contains the command-line applications for the Arbitrum sequencer decoder system.

## sequencer-reader

The sequencer-reader command connects to the Arbitrum RPC endpoint and reads transactions from the sequencer feed in real-time. It decodes DEX transactions and passes them to the arbitrage engine for analysis.

### Usage

```bash
go run cmd/sequencer-reader/main.go [flags]
```

### Flags

- `-rpc`: Arbitrum RPC endpoint (default: "https://arb1.arbitrum.io/rpc")
- `-sequencer`: Sequencer URL (optional)
- `-log-level`: Logging level (default: "info")
- `-config`: Configuration file path (optional)

## sequencer-capture

The sequencer-capture command captures transactions from the Arbitrum network and stores them to a file for later analysis or replay testing.

### Usage

```bash
go run cmd/sequencer-capture/main.go [flags]
```

### Flags

- `-rpc`: Arbitrum RPC endpoint (default: "https://arb1.arbitrum.io/rpc")
- `-output`: Output file for captured transactions (default: "testdata/sequencer/raw_transactions.jsonl")
- `-duration`: Duration to capture transactions (default: 1h, 0 for indefinite)
- `-filter-to`: Filter transactions to specific contract addresses (comma-separated)