# Transaction Classifier

The classifier package identifies DEX transactions by address and function signature. It maintains known addresses for various DEXes on Arbitrum and maps function signatures to specific operations.

## Features

- Identifies transactions to known DEX contracts
- Maps contract addresses to protocol names
- Identifies function types based on calldata signatures
- Supports major DEXes: Uniswap V3, Camelot V3, Curve, Balancer V2, Ramses, Kyber

## Usage

```go
package main

import (
    "fmt"
    "github.com/ethereum/go-ethereum/common"
    "arbitrum-sequencer-decoder/pkg/classifier"
)

func main() {
    // Example: classify a transaction
    toAddress := common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564") // Uniswap V3
    calldata := common.Hex2Bytes("791ac947000000000000000000000000...") // exactInput function signature
    
    protocol, functionType, isDEX := classifier.ClassifyTransaction(toAddress, calldata)
    if isDEX {
        fmt.Printf("DEX: %s, Function: %s\n", protocol, functionType)
    }
}
```

## Supported DEXes and Functions

The classifier currently supports:

- Uniswap V3
  - exactInput
  - exactOutput
- Camelot V3
  - swapExactTokensForTokens
- Curve
  - exchange
- Balancer V2
  - swap
- Ramses
  - swapExactAmounts
- Kyber
  - swap

## Adding New DEX Support

To add support for a new DEX:

1. Add the contract address to `KnownDEXAddresses` map
2. Add function signatures to `FunctionSignatures` map
3. Update documentation to reflect new support