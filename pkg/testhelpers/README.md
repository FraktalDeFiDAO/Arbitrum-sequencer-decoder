# Test Helpers

The testhelpers package provides utilities and helpers for testing the arbitrum-sequencer-decoder. These functions simplify the creation of test data and common testing operations.

## Features

- Helper functions to create mock transactions
- Functions to create test pools, tokens, and states
- Utilities for reading JSONL test data files
- Helper functions for creating test decoded actions and arbitrage opportunities

## Usage

```go
package main

import (
    "math/big"
    "github.com/ethereum/go-ethereum/common"
    "arbitrum-sequencer-decoder/pkg/testhelpers"
    "arbitrum-sequencer-decoder/pkg/types"
)

func main() {
    // Create a test token
    token := testhelpers.CreateTestToken(
        common.HexToAddress("0x123..."),
        "WETH",
        "Wrapped Ether",
        18,
    )
    
    // Create a test pool
    pool := testhelpers.CreateTestPool(
        common.HexToAddress("0x456..."),
        token,  // token0
        testhelpers.CreateTestToken(
            common.HexToAddress("0x789..."),
            "USDC",
            "USD Coin",
            6,
        ),  // token1
        types.UniswapV3Pool,
    )
    
    // Create a test transaction
    tx := testhelpers.CreateTestTransaction(
        common.HexToAddress("0xabc..."),
        []byte("0x791ac947..."), // calldata
        big.NewInt(0), // value
    )
}
```

## Available Functions

### `CreateTestTransaction(to, data, value) *types.Transaction`
Creates a mock Ethereum transaction for testing with the specified parameters.

### `CreateTestPool(address, token0, token1, poolType) *types.Pool`
Creates a test pool with specified parameters.

### `CreateTestToken(address, symbol, name, decimals) types.Token`
Creates a test token with specified parameters.

### `CreateTestPoolState(reserves0, reserves1) *types.PoolState`
Creates a test pool state with specified reserves and calculated prices.

### `ReadJSONLFile(filename, objType) []interface{}, error`
Reads a JSONL (JSON Lines) file and returns the parsed objects.

### `CreateTestDecodedAction(actionType, protocol, pool, tokenIn, tokenOut, amountIn, amountOut) types.DecodedAction`
Creates a test decoded action for testing.

### `CreateTestArbitrageOpportunity(oppType, profit, profitToken) types.ArbitrageOpportunity`
Creates a test arbitrage opportunity for testing.

## Integration Testing

These helpers are especially useful for integration testing where you need to set up complex scenarios with multiple components:

1. Use helpers to create consistent test data
2. Set up pools with known states
3. Create transactions that will trigger specific decoder behaviors
4. Validate outputs from the arbitrage engine

## Best Practices

- Use these helpers to maintain consistency across tests
- Create test data that represents realistic scenarios
- Use meaningful addresses and values for better debugging
- Combine helpers to create complex test scenarios