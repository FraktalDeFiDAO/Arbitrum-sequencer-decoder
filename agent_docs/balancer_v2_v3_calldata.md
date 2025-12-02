# Balancer V2/V3 Calldata Decoding

## Purpose
Decode pending swap transactions targeting Balancer V2/V3 vaults from raw Arbitrum sequencer data. Balancer pools use **weighted math** (V2) and **dynamic fees** (V3) which require different approaches to price impact estimation compared to constant-product AMMs.

## Relevant Source
- **Go ABI access**: `pkg/abi/balancer/`
  - `vault.go` → `VaultABI`
  - `weighted_pool.go` → `WeightedPoolABI`
  - `stable_pool.go` → `StablePoolABI`
  - `composable_stable_pool.go` → `ComposableStablePoolABI`
- **Raw ABIs**: `./abis/balancer/vault.json`, `./abis/balancer/weighted_pool.json`, etc.
- Decoder: `pkg/decoder/balancer.go`
- Simulator: `pkg/simulator/balancer/`

## Key Functions in Calldata
Balancer V2/V3 uses a single `swap` entry point with detailed swap structs:
- `swap((address poolId, address assetIn, address assetOut, uint256 amount, bytes userData), address from, address to, uint256 deadline, uint256 value)`
- `batchSwap(SwapKind kind, BatchSwapStep[] swaps, address[] assets, FundManagement funds, int256[] limits, uint256 deadline)`
- `queryBatchSwap(SwapKind kind, BatchSwapStep[] swaps, address[] assets, FundManagement funds)` - static call for estimation

> **Note**: `poolId` is a 32-byte identifier (keccak256 hash of pool parameters)

## Pool Resolution and Identification
1. **Vault address** on Arbitrum: `0xBA12222222228d8Ba445958a75a0704d566BF2C8`
2. To get pool state:
   - Use `VaultABI` to call `getPoolTokens(poolId)` to get:
     - `address[] tokens` - array of token addresses
     - `uint256[] balances` - current balances of each token
     - `uint256[] lastChangeBlock` - for TWAP calculations
3. For pool type: Use `VaultABI` call `getPool(poolId)` to get pool address, then check pool type via `IStablePool` or `IWeightedPool` interfaces

## Price Estimation: Weighted Pools (V2)
- Generalized AMM: `∏(token_balance[i] / weight[i]) = invariant`
- Price between two tokens: `price = (balance_token2 / weight_token2) / (balance_token1 / weight_token1)`
- Swap impact: Calculate new balances using weights and solve for invariant

## Price Estimation: Stable Pools (V2/V3)
- Use Stable Math formula: `A * n^n * ∏x_i + D = D * A * n^n + D^{n+1}/(n^n * ∏x_i)`
- Where `A` is amplification coefficient (available via pool ABI)
- Similar to Curve but with different implementation details

## Batch Swap Decoding
- Extract `BatchSwapStep[]` from calldata to identify all swaps in single transaction
- Each swap may affect subsequent ones in the batch due to balance updates
- For accurate pricing, process batch swaps sequentially considering balance changes

## Testing
```bash
go test ./pkg/decoder -run TestBalancerDecoder -v
go test ./pkg/simulator/balancer -v
```

## Performance Considerations
- Balancer pools have more complex math than Uniswap - cache invariant calculations
- Stable pools require iterative solving - implement efficient convergence
- Batch swaps can contain multiple hops - optimize for single calculation pass