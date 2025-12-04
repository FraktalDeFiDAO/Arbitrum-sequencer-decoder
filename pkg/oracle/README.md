# Pool Oracle

The oracle package tracks in-memory pool states for efficient access and updates. It provides methods to retrieve, update, and manage liquidity pool information necessary for price calculations and arbitrage detection.

## Features

- Concurrent access protection using read-write mutexes
- In-memory storage for fast access
- Support for both pool metadata and state data
- Thread-safe operations
- Pool tracking and management

## Core Components

### PoolOracle
The main struct that manages all pool information with methods for:
- Retrieving pool information
- Updating pool data
- Managing pool states
- Checking pool tracking status

## Usage

```go
package main

import (
    "math/big"
    "github.com/ethereum/go-ethereum/common"
    "arbitrum-sequencer-decoder/pkg/oracle"
    "arbitrum-sequencer-decoder/pkg/types"
)

func main() {
    // Create a new oracle instance
    oracle := oracle.NewPoolOracle()

    // Create a sample pool
    poolAddress := common.HexToAddress("0x123...")
    pool := &types.Pool{
        Address: poolAddress,
        Type: types.UniswapV3Pool,
        // ... other fields
    }

    // Update the oracle with the pool
    oracle.UpdatePools([]*types.Pool{pool})

    // Retrieve the pool
    retrievedPool, err := oracle.GetPool(poolAddress)
    if err != nil {
        // Handle error
    }

    // Update the pool state
    state := &types.PoolState{
        Reserves0: big.NewInt(1000000),
        Reserves1: big.NewInt(2000000),
        // ... other state fields
    }
    oracle.UpdatePoolState(poolAddress, state)
}
```

## Methods

### `NewPoolOracle() *PoolOracle`
Creates a new PoolOracle instance.

### `GetPool(address common.Address) (*types.Pool, error)`
Returns the current state of a pool.

### `UpdatePools(pools []*types.Pool) error`
Updates multiple pools with new data.

### `GetPoolState(address common.Address) (*types.PoolState, error)`
Returns the detailed state of a pool.

### `UpdatePoolState(address common.Address, state *types.PoolState) error`
Updates the state of a specific pool.

### `RemovePool(address common.Address) error`
Removes a pool from the oracle.

### `GetAllPools() []*types.Pool`
Returns all tracked pools.

### `GetAllPoolStates() map[common.Address]*types.PoolState`
Returns all tracked pool states.

### `IsPoolTracked(address common.Address) bool`
Checks if a pool is being tracked by the oracle.

### `GetRecentPools(duration time.Duration) []*types.Pool`
Returns pools updated within the specified duration.

## Thread Safety

The PoolOracle is thread-safe and can be safely accessed from multiple goroutines. Read operations use read locks, while write operations use full locks to ensure data consistency.