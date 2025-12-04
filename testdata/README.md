# Test Data for Arbitrum Sequencer Decoder

This directory contains test data used for validating the Arbitrum sequencer decoder system. The test data is organized to support various aspects of the system including:

## Directory Structure
```
testdata/
├── sequencer/           # Raw sequencer transaction data
│   ├── raw_transactions.jsonl    # Captured raw transactions from sequencer
│   ├── uniswap_v3_swaps.jsonl    # Uniswap V3 swap transactions
│   ├── camelot_v3_swaps.jsonl    # Camelot V3 swap transactions
│   ├── curve_swaps.jsonl         # Curve swap transactions
│   └── mixed_dex_transactions.jsonl # Mixed DEX transactions for integration
└── historical/          # Historical data for replay testing
    ├── arbitrage_opportunities.jsonl # Previously profitable opportunities
    └── backtesting_data.jsonl         # Data for strategy backtesting
```

## Data Formats

### JSONL Format
Most files in this directory use JSONL (JSON Lines) format, where each line is a valid JSON object. This allows for streaming processing and easy parsing.

Example:
```json
{"hash":"0x123...","block_number":12345,"timestamp":"2023-01-01T00:00:00Z","from":"0x...","to":"0x...","data":"0x..."}
```

## Usage in Testing

### Unit Testing
- Individual decoder tests use specific transaction samples
- Price simulation tests use historical pool state data
- Error condition testing uses malformed transaction examples

### Integration Testing
- Mixed DEX transaction files test the full pipeline
- Sequencer replay tests validate end-to-end functionality

### Performance Testing
- Large transaction files test processing throughput
- Stress testing uses high-volume transaction sequences

## Maintaining Test Data

When adding new DEX support:
1. Add sample transactions to the appropriate sequencer file
2. Include both successful and error condition examples
3. Document any special handling required in comments
4. Ensure transactions represent real-world usage patterns