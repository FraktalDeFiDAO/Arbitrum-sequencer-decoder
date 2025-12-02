# Curve Pool Solver

## Purpose
Estimate post-swap price impact for Curve pools (stableswap and crypto) using **pure Go math**—no EVM simulation—based on raw sequencer calldata and current on-chain state.

## Relevant Source
- **Go ABI access**: `pkg/abi/curve/`
  - `factory_plain.go`, `factory_crypto.go` → factory ABIs
  - `pool_stable.go`, `pool_crypto.go` → pool ABIs
- **Raw ABIs**: `./abis/curve/factory-plain.json`, `./abis/curve/pool-stable.json`, etc.
- Solver logic: `pkg/simulator/curve/`
- Decoder: `pkg/decoder/curve.go`

## How to Resolve a Curve Pool
Curve has multiple factories. To find the correct pool for a token pair:
1. **For stablecoins (e.g., USDC/USDT/DAI)**:
   - Query `StableSwapFactory` (`0x9AFf31761625314A7110BA401C6F7319cF126e43` on Arbitrum) via `find_pool_for_tokens(tokenA, tokenB)`
2. **For crypto pairs (e.g., WETH/WBTC)**:
   - Query `CryptoSwapFactory` (`0x9D0464996170c6B9e75eED71c68B99dDEDf279e8`) via `find_pool_for_coins(tokenA, tokenB)`
3. If no pool found, check **registry** (`0x971E732Bca93F7663461B66fC330D47c9710357C`) as fallback

> 🔑 **Never assume pool addresses**. Always resolve via factory or registry.

## Key On-Chain Values (Fetch via ABI)
Once you have the pool address, read:
- **Stableswap**:
  - `coins()` → [token0, token1, ...]
  - `balances()` → [uint256]
  - `A()` → amplification coefficient
  - `get_dy(i, j, dx)` → estimate output (used for price)
- **Cryptoswap**:
  - `coins()`, `A()`, `gamma()`
  - `price_oracle()` → smoothed price
  - Use `get_dy` or invariant derivative for spot price

All calls are **static**—no state changes.

## Price Impact Estimation
- **Stableswap**: Solve `A * n^n * ∏x_i + D = D * A * n^n + D^{n+1}/(n^n * ∏x_i)` iteratively in Go.
- **Cryptoswap**: Use analytical derivative of `K = γ * (A * χ + 1) * ∏x_i` where `χ = (∑x_i)^2 / (n * ∏x_i)`
- Implementations in `pkg/simulator/curve/stableswap.go` and `cryptoswap.go`

> ⚠️ **No logs in sequencer feed**. Only calldata is available—decode `exchange(i, j, dx, min_dy)` or `swap` calls directly.

## Testing
```bash
go test ./pkg/simulator/curve -v