# Kyber Classic/Elastic Calldata Decoding

## Purpose
Decode pending swap transactions targeting Kyber Classic (Standard, Dynamic) and Kyber Elastic (Kandel) contracts on Arbitrum. Kyber uses **hybrid liquidity** with concentrated liquidity positions in Elastic and dynamic fees in Classic variants.

## Relevant Source
- **Go ABI access**: `pkg/abi/kyber/`
  - `elastic_router.go` → `ElasticRouterABI`
  - `classic_router.go` → `ClassicRouterABI`
  - `pair.go` → `PairABI`
  - `pool.go` → `PoolABI`
- **Raw ABIs**: `./abis/kyber/elastic-router.json`, `./abis/kyber/classic-router.json`, `./abis/kyber/pair.json`
- Decoder: `pkg/decoder/kyber.go`
- Simulator: `pkg/simulator/kyber/`

## Key Functions in Calldata

### Kyber Classic Router Functions:
- `swapExactTokensForTokens(uint256 amountIn, uint256 amountOutMin, address[] path, address to, bytes referrer)`
- `swapExactETHForTokens(uint256 amountOutMin, address[] path, address to, bytes referrer)`
- `swapTokensForExactTokens(uint256 amountOut, uint256 amountInMax, address[] path, address to, bytes referrer)`

### Kyber Elastic Router Functions:
- `exactInputSingle((address tokenIn, address tokenOut, uint24 fee, address recipient, uint256 deadline, uint256 amountIn, uint256 amountOutMinimum, uint160 sqrtPriceLimitX96))`
- `exactOutputSingle((address tokenIn, address tokenOut, uint24 fee, address recipient, uint256 deadline, uint256 amountOut, uint256 amountInMaximum, uint160 sqrtPriceLimitX96))`
- `exactInput((bytes path, address recipient, uint256 deadline, uint256 amountIn, uint256 amountOutMinimum))`

## Pool Resolution
1. **Classic Router address** on Arbitrum: `0x613B1F38C75A5C3E564f7c886Fd39e38e62b3e69`
2. **Elastic Router address** on Arbitrum: `0x2B1c7b41f6A8F2b2bc45C3b34e84cfD8b097B6B4`
3. **Factory addresses**:
   - Classic: `0xC7a15f4dcb6F34266D9477571b21C2D0aC35e73c`
   - Elastic: `0x2Ce32033838d01Ee79a0a0727cB41d7a3d3D6160`

To get pair/pool state:
- For Classic: Use factory `getPair(tokenA, tokenB)` to get pair address
- For Elastic: Use factory `getPool(tokenA, tokenB, feeTier)` to get pool address
- Read reserves or tick state depending on pool type

## Price Estimation: Kyber Classic
- Uses modified constant product: `xy = k` with dynamic fees
- Reserve-based pricing: `price = reserveOut / reserveIn` 
- Dynamic fee calculation available via `getReservesAndFee` method
- Fee range: typically 0.04% to 1.0% depending on volatility

## Price Estimation: Kyber Elastic (CLMM)
- Uses concentrated liquidity concept similar to Uniswap V3
- Price determined by current tick and active liquidity
- Use tick math: `price = 1.0001^tick`
- Calculate price impact across ticks with active liquidity

## Calldata Parsing Strategy
1. **Verify router address** matches known Kyber contracts
2. **Parse function selector** to identify swap type
3. **Decode parameters** using `go-ethereum/accounts/abi`
4. **Extract token path** and amounts for multi-hop swaps
5. **Calculate route-specific pricing** for each path segment

## Testing
```bash
go test ./pkg/decoder -run TestKyberDecoder -v
go test ./pkg/simulator/kyber -v
```

## Special Considerations
- Kyber Classic includes referrer tracking - decode referrer without affecting pricing
- Elastic pools have tick-based range orders - consider active liquidity when pricing
- Dynamic fees in Classic require additional chain calls for accurate estimation
- Elastic uses sqrtPriceX96 format - convert to readable price for simulation