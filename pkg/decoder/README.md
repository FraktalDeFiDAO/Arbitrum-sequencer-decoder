# DEX Transaction Decoder Package

The decoder package provides interfaces and base functionality for decoding DEX transaction calldata. Each DEX protocol has its own decoder that implements the common Decoder interface.

## Architecture

The package includes:

- `interface.go`: Defines the common Decoder interface and BaseDecoder
- Individual decoder directories for each DEX protocol
- Utility functions for common decoding tasks

## Common Decoder Interface

All DEX decoders implement the Decoder interface with three methods:

1. `Matches(tx, toAddress)`: Determines if the decoder should handle a transaction
2. `Decode(tx, toAddress)`: Decodes transaction calldata into DecodedAction(s)
3. `Protocol()`: Returns the protocol type

## BaseDecoder

The BaseDecoder provides common fields and functionality that all decoders can use.

## Utility Functions

- `MatchesSignature(tx, signatures)`: Checks if transaction calldata matches any of the provided function signatures

## Implemented Decoders

### DEX Protocols
| Protocol | Directory | Status | Supported Functions |
|----------|-----------|--------|---------------------|
| Uniswap V3 | `uniswap_v3/` | Complete | exactInput, exactOutput, exactInputSingle, exactOutputSingle, multicall |
| Camelot V3 | `camelot_v3/` | Complete | swapExactTokensForTokens, swapTokensForExactTokens |
| SushiSwap | `sushiswap/` | Complete | processRoute, swapExactTokensForTokens, swapExactETHForTokens, swapExactTokensForETH |
| KyberSwap | `kyberswap/` | Complete | swap, swapSimpleMode, swapGeneric |

### Aggregators
| Protocol | Directory | Status | Supported Functions |
|----------|-----------|--------|---------------------|
| 1inch | `oneinch/` | Complete | swap, unoswap, unoswapTo, uniswapV3Swap, uniswapV3SwapTo |
| OpenOcean | `openocean/` | Complete | swap, callUniswapTo |
| Odos | `odos/` | Complete | swap, swapMulti, swapRouterTokensForTokens |

### Planned
- Curve
- Balancer V2/V3
- Ramses
- GMX
- Trader Joe
- WooFi

## Adding a New Decoder

1. Create a new directory under `pkg/decoder/<protocol_name>/`
2. Implement `abi_bindings.go` with:
   - Protocol ABI as a constant string
   - Known router addresses map
   - Function signatures list and map
3. Implement `decoder.go` with:
   - Decoder struct embedding `BaseDecoder`
   - `NewXxxDecoder()` constructor
   - `Matches()`, `Decode()`, `Protocol()` methods
   - Individual decode functions for each supported method
4. Add the decoder to `cmd/sequencer-reader/main.go`
5. Add router addresses to `pkg/classifier/classifier.go`

## Example Usage

```go
// Initialize decoder
decoder := uniswap_v3.NewUniswapV3Decoder()

// Check if decoder matches transaction
if decoder.Matches(tx, toAddress) {
    actions, err := decoder.Decode(tx, toAddress)
    if err != nil {
        log.Printf("Decode error: %v", err)
        continue
    }
    for _, action := range actions {
        log.Printf("Swap: %s -> %s, Amount: %s",
            action.TokenIn.Address.Hex(),
            action.TokenOut.Address.Hex(),
            action.AmountIn.String())
    }
}
```