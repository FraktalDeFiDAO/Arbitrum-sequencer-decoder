# Ramses Calldata Decoding

## Purpose
Decode pending swap transactions targeting Ramses V2 and V3 contracts on Arbitrum. Ramses is a fork of Uniswap V3 with additional features including **veRAM governance** and dynamic fee structures that influence price impact calculations.

## Relevant Source
- **Go ABI access**: `pkg/abi/ramses/`
  - `router.go` → `RouterABI`
  - `factory.go` → `FactoryABI`
  - `pool.go` → `PoolABI`
- **Raw ABIs**: `./abis/ramses/router.json`, `./abis/ramses/factory.json`, `./abis/ramses/pool.json`
- Decoder: `pkg/decoder/ramses.go`
- Simulator: `pkg/simulator/ramses/`

## Key Functions in Calldata
Ramses V3 supports similar functions to Uniswap V3 with added governance features:
- `exactInput((bytes path, uint256 deadline, uint256 amountIn, uint256 amountOutMinimum, address recipient))`
- `exactOutput((bytes path, uint256 deadline, uint256 amountOut, uint256 amountInMaximum, address recipient))`
- `exactInputSingle((address tokenIn, address tokenOut, uint256 deadline, uint256 amountIn, uint256 amountOutMinimum, address recipient, uint160 sqrtPriceLimitX96))`
- `exactOutputSingle((address tokenIn, address tokenOut, uint256 deadline, uint256 amountOut, uint256 amountInMaximum, address recipient, uint160 sqrtPriceLimitX96))`

## Pool Resolution
1. **Router address** on Arbitrum: `0x792D76a3De6Bd9eF3c32988b7e9C57dB30779b51`
2. **Factory address** on Arbitrum: `0xAAA87963EFeB6f9b7e2b837B3Aefe20300252Ec7`
3. To get pool state:
   - Use `FactoryABI` to call `getPool(tokenA, tokenB, feeTier)` to get pool address
   - Ramses uses same fee tiers as Uniswap V3: 100, 500, 3000, 10000
   - Read tick state, liquidity, and oracle data from pool contract

## Price Estimation
- Ramses uses identical tick math to Uniswap V3: `price = 1.0001^tick`
- However, veRAM governance may influence fee structures
- Liquidity calculations follow Uniswap V3 concentrated liquidity model
- Consider dynamic fees when available through governance contracts

## Special Ramses Features
- **Protocol Fee**: Ramses may implement additional protocol fees beyond LP fees
- **veRAM Integration**: Governance tokens may affect fee distribution
- **Range Orders**: Similar to Uniswap V3 concentrated positions
- **Cross-chain Integration**: Potential cross-chain fee sharing mechanisms

## Calldata Parsing Strategy
1. **Match router address** against known Ramses router contracts
2. **Decode function selector** to identify exactInput/Output variants
3. **Parse path encoding** same as Uniswap V3 (token + fee bytes)
4. **Estimate price impact** using Uniswap V3 tick math (with Ramses-specific parameters)
5. **Factor in any protocol fees** from governance mechanisms

## Testing
```bash
go test ./pkg/decoder -run TestRamsesDecoder -v
go test ./pkg/simulator/ramses -v
```

## Performance Considerations
- Ramses pools have same computational requirements as Uniswap V3
- Cache tick information for frequently accessed pools
- Consider governance data impacts on fee calculations for accuracy
- Ramses follows Uniswap V3 efficiency patterns for large transaction volumes