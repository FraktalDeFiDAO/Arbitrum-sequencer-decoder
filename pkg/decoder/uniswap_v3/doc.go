// Package uniswap_v3 provides a complete implementation for decoding Uniswap V3 transactions
package uniswap_v3

// This package provides:
// 1. ABI bindings for Uniswap V3 router functions
// 2. A decoder that implements the Decoder interface
// 3. Path decoding for multi-hop swaps
// 4. Pool state tracking
// 5. Price simulation functions
//
// The implementation follows the architecture defined in the project plan:
// - Uses pure Go for mathematical calculations (no EVM required)
// - Efficiently decodes complex path routing
// - Handles all Uniswap V3 function types (exactInput, exactOutput, etc.)
