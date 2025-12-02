# Uniswap V3 Calldata Decoding

## Purpose
Decode pending swap transactions targeting Uniswap V3 routers (`SwapRouter`, `UniversalRouter`) from raw Arbitrum sequencer data **before execution**. Logs and receipts are unavailable—only calldata and recipient address are present.

## Relevant Source
- `pkg/decoder/uniswap_v3.go`
- `pkg/simulator/uniswap_v3_math.go`
- ABIs: `./abis/uniswap/v3/SwapRouter.json`, `./abis/uniswap/v3/Quoter.json`

## Key Functions in Calldata
Uniswap V3 supports two primary swap entry points:
1. `exactInput((bytes path, uint24 deadline, uint256 amountIn, uint256 amountOutMinimum, address recipient))`
2. `exactOutput((bytes path, uint24 deadline, uint256 amountOut, uint256 amountInMaximum, address recipient))`

Both use a **compact path encoding**:
- Each hop is: `tokenIn (20 bytes) + fee (3 bytes)`
- Final token: `tokenOut (20 bytes)`
- Example path (USDC → WETH @ 0.3%):  
  `0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48` + `0x000bb8` + `0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2`

> ⚠️ **No event parsing**: `Swap` events fire *after* execution. Only calldata is available from sequencer.

## Decoding Steps (Go)
1. Match `to` address against known Uniswap V3 router addresses on Arbitrum:
   - `0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45` (SwapRouter)
   - `0x3fC91A3afd70395Cd496C647d5a6CC9D4B2b7FAD` (UniversalRouter)
2. Parse 4-byte function selector:
   - `0x41556510` → `exactInput`
   - `0x122f42f8` → `exactOutput`
3. ABI-decode the tuple using `go-ethereum/accounts/abi`
4. Extract:
   - `amountIn` or `amountOut`
   - `path` (decode into token/fee hops)
   - `recipient`
5. Estimate post-swap price using **stateless tick math** (see `pkg/simulator/uniswap_v3_math.go`)

## Price Impact Estimation
Use analytical Uniswap V3 formulas:
- Given current tick, liquidity, and amount, compute new tick and sqrtPriceX96.
- Do **not** simulate with EVM—use pure Go functions from `uniswap/v3-core` port.

## Testing
Replay sequencer samples:
```bash
go test ./pkg/decoder -run TestUniswapV3Decoder -v

