# Camelot V3 Router Decoding

## Purpose
Decode pending swaps targeting Camelot V3 router on Arbitrum from raw sequencer calldata. Unlike Uniswap V2, Camelot V3 includes **LP fee hooks**, **reward emissions**, and **custom path encoding**—all of which must be handled during pre-execution analysis.

## Relevant Source
- **Go ABI access**: `pkg/abi/camelot/v3/`
  - `factory.go` → `FactoryABI`
  - `pair.go` → `PairABI`
  - `router.go` → `RouterABI`
- **Raw ABIs**: `./abis/camelot/v3/Factory.json`, `./abis/camelot/v3/Pair.json`, `./abis/camelot/v3/Router.json`
- Decoder: `pkg/decoder/camelot_v3.go`
- Simulator: `pkg/simulator/camelot_v3.go`

## Key Router Functions (Calldata Selectors)
Camelot V3 uses a modified Uniswap V2 interface with additional hooks:
- `swapExactTokensForTokens(uint256 amountIn, uint256 amountOutMin, address[] path, address to, uint256 deadline)`
- `swapExactETHForTokens(uint256 amountOutMin, address[] path, address to, uint256 deadline)`
- `swapExactTokensForETH(uint256 amountIn, uint256 amountOutMin, address[] path, address to, uint256 deadline)`

> Function selectors differ from Uniswap V2 due to custom bytecode. Verify against `RouterABI`.

## Pool & Factory Resolution
1. **Router address** on Arbitrum: `0x4F9254C83EB525f9FCf346490bbb3ed28a81c667`
2. To get pool state:
   - Use `FactoryABI` to call `getPair(tokenA, tokenB)` on factory (`0x9c3767645F2016eB6b2473f611aB02b62C95C9Ae`)
   - Camelot V3 **does not use fee tiers**—each pair has one canonical pool
3. With pair address, use `PairABI` to read:
   - `getReserves()` → `(uint112 reserve0, uint112 reserve1, uint32 blockTimestampLast)`
   - `totalSupply()` → for LP-based reward estimation (not needed for arbitrage)

> ⚠️ **No event reliance**: `Swap` events are post-execution. Only calldata is available from sequencer.

## Price Estimation
- Use constant product formula: `price = reserve1 / reserve0` (adjusted for decimals)
- Account for **fee-on-transfer tokens**? → No. Arbitrage engine assumes standard ERC20s. Non-compliant tokens are excluded via allowlist in `pkg/oracle/token_filter.go`
- Simulate swap impact using:  
  `delta_y = (reserve1 * amountIn * 997) / (reserve0 * 1000 + amountIn * 997)`  
  (0.3% fee baked into math)

## Testing
```bash
go test ./pkg/decoder -run TestCamelotV3Decoder -v
