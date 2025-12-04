# Uniswap V3 Decoder

This package provides decoding functionality for Uniswap V3 transactions on the Arbitrum network.

## Components

- `abi_bindings.go`: Contains the ABI definitions for Uniswap V3 router functions
- `decoder.go`: Implements the Uniswap V3 transaction decoder
- `path_decoder.go`: Handles Uniswap V3 path decoding functionality

## Supported Functions

- `exactInput`: Swap a fixed amount of input tokens for a variable amount of output tokens across a path
- `exactOutput`: Swap a variable amount of input tokens for a fixed amount of output tokens across a path
- `exactInputSingle`: Swap a fixed amount of one token for another token directly
- `exactOutputSingle`: Swap a variable amount of one token for a fixed amount of another token directly

## Features

- Path decoding for multi-hop swaps
- Support for concentrated liquidity positions
- Fee tier extraction from path data
- Integration with the base decoder interface