# Camelot V3 Decoder

This package provides decoding functionality for Camelot V3 transactions on the Arbitrum network.

## Components

- `abi_bindings.go`: Contains the ABI definitions for Camelot V3 router functions
- `decoder.go`: Implements the Camelot V3 transaction decoder
- `decoder_test.go`: Tests for the Camelot V3 decoder functionality

## Supported Functions

- `swapExactTokensForTokens`: Swap a fixed amount of input tokens for a variable amount of output tokens
- `swapTokensForExactTokens`: Swap a variable amount of input tokens for a fixed amount of output tokens
- `swapExactTokensForTokensSupportingFeeOnTransferTokens`: Swap supporting fee-on-transfer tokens

## Features

- Integration with the base decoder interface
- Support for complex swap parameters
- Volume tracking for liquidity calculations