// Package uniswap_v3 implements decoder for Uniswap V3 transactions
package uniswap_v3

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// SwapRouter02 ABI - has different param structures (no deadline in struct)
const Router02ABI = `[
{
    "inputs": [
        {
            "components": [
                {"internalType": "address", "name": "tokenIn", "type": "address"},
                {"internalType": "address", "name": "tokenOut", "type": "address"},
                {"internalType": "uint24", "name": "fee", "type": "uint24"},
                {"internalType": "address", "name": "recipient", "type": "address"},
                {"internalType": "uint256", "name": "amountIn", "type": "uint256"},
                {"internalType": "uint256", "name": "amountOutMinimum", "type": "uint256"},
                {"internalType": "uint160", "name": "sqrtPriceLimitX96", "type": "uint160"}
            ],
            "internalType": "struct IV3SwapRouter.ExactInputSingleParams",
            "name": "params",
            "type": "tuple"
        }
    ],
    "name": "exactInputSingle",
    "outputs": [{"internalType": "uint256", "name": "amountOut", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
},
{
    "inputs": [
        {
            "components": [
                {"internalType": "bytes", "name": "path", "type": "bytes"},
                {"internalType": "address", "name": "recipient", "type": "address"},
                {"internalType": "uint256", "name": "amountIn", "type": "uint256"},
                {"internalType": "uint256", "name": "amountOutMinimum", "type": "uint256"}
            ],
            "internalType": "struct IV3SwapRouter.ExactInputParams",
            "name": "params",
            "type": "tuple"
        }
    ],
    "name": "exactInput",
    "outputs": [{"internalType": "uint256", "name": "amountOut", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
},
{
    "inputs": [
        {
            "components": [
                {"internalType": "address", "name": "tokenIn", "type": "address"},
                {"internalType": "address", "name": "tokenOut", "type": "address"},
                {"internalType": "uint24", "name": "fee", "type": "uint24"},
                {"internalType": "address", "name": "recipient", "type": "address"},
                {"internalType": "uint256", "name": "amountOut", "type": "uint256"},
                {"internalType": "uint256", "name": "amountInMaximum", "type": "uint256"},
                {"internalType": "uint160", "name": "sqrtPriceLimitX96", "type": "uint160"}
            ],
            "internalType": "struct IV3SwapRouter.ExactOutputSingleParams",
            "name": "params",
            "type": "tuple"
        }
    ],
    "name": "exactOutputSingle",
    "outputs": [{"internalType": "uint256", "name": "amountIn", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
},
{
    "inputs": [
        {
            "components": [
                {"internalType": "bytes", "name": "path", "type": "bytes"},
                {"internalType": "address", "name": "recipient", "type": "address"},
                {"internalType": "uint256", "name": "amountOut", "type": "uint256"},
                {"internalType": "uint256", "name": "amountInMaximum", "type": "uint256"}
            ],
            "internalType": "struct IV3SwapRouter.ExactOutputParams",
            "name": "params",
            "type": "tuple"
        }
    ],
    "name": "exactOutput",
    "outputs": [{"internalType": "uint256", "name": "amountIn", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
}
]`

// Uniswap V3 Router ABI for swap functions (original SwapRouter)
// This is the correct ABI from the official Uniswap V3 SwapRouter contract
const RouterABI = `[
{
    "inputs": [
        {
            "components": [
                {"internalType": "bytes", "name": "path", "type": "bytes"},
                {"internalType": "address", "name": "recipient", "type": "address"},
                {"internalType": "uint256", "name": "deadline", "type": "uint256"},
                {"internalType": "uint256", "name": "amountIn", "type": "uint256"},
                {"internalType": "uint256", "name": "amountOutMinimum", "type": "uint256"}
            ],
            "internalType": "struct ISwapRouter.ExactInputParams",
            "name": "params",
            "type": "tuple"
        }
    ],
    "name": "exactInput",
    "outputs": [{"internalType": "uint256", "name": "amountOut", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
},
{
    "inputs": [
        {
            "components": [
                {"internalType": "address", "name": "tokenIn", "type": "address"},
                {"internalType": "address", "name": "tokenOut", "type": "address"},
                {"internalType": "uint24", "name": "fee", "type": "uint24"},
                {"internalType": "address", "name": "recipient", "type": "address"},
                {"internalType": "uint256", "name": "deadline", "type": "uint256"},
                {"internalType": "uint256", "name": "amountIn", "type": "uint256"},
                {"internalType": "uint256", "name": "amountOutMinimum", "type": "uint256"},
                {"internalType": "uint160", "name": "sqrtPriceLimitX96", "type": "uint160"}
            ],
            "internalType": "struct ISwapRouter.ExactInputSingleParams",
            "name": "params",
            "type": "tuple"
        }
    ],
    "name": "exactInputSingle",
    "outputs": [{"internalType": "uint256", "name": "amountOut", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
},
{
    "inputs": [
        {
            "components": [
                {"internalType": "bytes", "name": "path", "type": "bytes"},
                {"internalType": "address", "name": "recipient", "type": "address"},
                {"internalType": "uint256", "name": "deadline", "type": "uint256"},
                {"internalType": "uint256", "name": "amountOut", "type": "uint256"},
                {"internalType": "uint256", "name": "amountInMaximum", "type": "uint256"}
            ],
            "internalType": "struct ISwapRouter.ExactOutputParams",
            "name": "params",
            "type": "tuple"
        }
    ],
    "name": "exactOutput",
    "outputs": [{"internalType": "uint256", "name": "amountIn", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
},
{
    "inputs": [
        {
            "components": [
                {"internalType": "address", "name": "tokenIn", "type": "address"},
                {"internalType": "address", "name": "tokenOut", "type": "address"},
                {"internalType": "uint24", "name": "fee", "type": "uint24"},
                {"internalType": "address", "name": "recipient", "type": "address"},
                {"internalType": "uint256", "name": "deadline", "type": "uint256"},
                {"internalType": "uint256", "name": "amountOut", "type": "uint256"},
                {"internalType": "uint256", "name": "amountInMaximum", "type": "uint256"},
                {"internalType": "uint160", "name": "sqrtPriceLimitX96", "type": "uint160"}
            ],
            "internalType": "struct ISwapRouter.ExactOutputSingleParams",
            "name": "params",
            "type": "tuple"
        }
    ],
    "name": "exactOutputSingle",
    "outputs": [{"internalType": "uint256", "name": "amountIn", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
},
{
    "inputs": [{"internalType": "bytes[]", "name": "data", "type": "bytes[]"}],
    "name": "multicall",
    "outputs": [{"internalType": "bytes[]", "name": "results", "type": "bytes[]"}],
    "stateMutability": "payable",
    "type": "function"
},
{
    "inputs": [
        {"internalType": "uint256", "name": "deadline", "type": "uint256"},
        {"internalType": "bytes[]", "name": "data", "type": "bytes[]"}
    ],
    "name": "multicall",
    "outputs": [{"internalType": "bytes[]", "name": "results", "type": "bytes[]"}],
    "stateMutability": "payable",
    "type": "function"
}
]`

// ABI instance for Uniswap V3 Router - use GetRouterABI() instead of direct access
var (
	routerABIOnce    sync.Once
	routerABI        abi.ABI
	routerABIInitErr error

	router02ABIOnce    sync.Once
	router02ABI        abi.ABI
	router02ABIInitErr error
)

// GetRouterABI returns the parsed Uniswap V3 Router ABI.
// Thread-safe and returns any initialization error instead of panicking.
func GetRouterABI() (abi.ABI, error) {
	routerABIOnce.Do(func() {
		routerABI, routerABIInitErr = abi.JSON(strings.NewReader(RouterABI))
	})
	return routerABI, routerABIInitErr
}

// GetRouter02ABI returns the parsed Uniswap V3 SwapRouter02 ABI.
// Thread-safe and returns any initialization error instead of panicking.
func GetRouter02ABI() (abi.ABI, error) {
	router02ABIOnce.Do(func() {
		router02ABI, router02ABIInitErr = abi.JSON(strings.NewReader(Router02ABI))
	})
	return router02ABI, router02ABIInitErr
}

// MustGetRouterABI returns the Router ABI or panics.
// Only use during application startup when ABI errors should be fatal.
func MustGetRouterABI() abi.ABI {
	abi, err := GetRouterABI()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize Uniswap V3 Router ABI: %v", err))
	}
	return abi
}

// RouterABIInstance is deprecated - use GetRouterABI() instead
// Kept for backwards compatibility
var RouterABIInstance abi.ABI

// Initialize the ABI instance (for backwards compatibility)
func init() {
	var err error
	RouterABIInstance, err = abi.JSON(strings.NewReader(RouterABI))
	if err != nil {
		// Log error but don't panic - use GetRouterABI() for safe access
		fmt.Printf("Warning: Failed to initialize RouterABIInstance: %v\n", err)
	}
}

// Known Uniswap V3 Router addresses on Arbitrum
var KnownRouterAddresses = map[common.Address]string{
	// Arbitrum Uniswap V3 SwapRouter (original)
	common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564"): "UniswapV3SwapRouter",
	// Arbitrum Uniswap V3 SwapRouter02 (newer version with different selectors)
	common.HexToAddress("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45"): "UniswapV3SwapRouter02",
	common.HexToAddress("0x0BFbCf88ED4e3dfCb819Cf5AD3b6730c35e6C378"): "UniswapV3Router2",
}

// Function signatures for Uniswap V3 swap operations
// Verified against Uniswap V3 SwapRouter contract on Arbitrum
var FunctionSignatures = []string{
	// SwapRouter (original) signatures
	"0xc04b8d59", // exactInput(ExactInputParams) - multi-hop exact input swap
	"0xf28c0498", // exactOutput(ExactOutputParams) - multi-hop exact output swap
	"0x414bf389", // exactInputSingle(ExactInputSingleParams) - single-hop exact input
	"0xdb3e2198", // exactOutputSingle(ExactOutputSingleParams) - single-hop exact output
	"0xac9650d8", // multicall(bytes[]) - batch multiple calls
	"0x5ae401dc", // multicall(uint256,bytes[]) - batch with deadline
	// SwapRouter02 specific signatures (different param struct, no deadline in params)
	"0x04e45aaf", // exactInputSingle (SwapRouter02 variant)
	"0xb858183f", // exactInput (SwapRouter02 variant)
	"0x09b81346", // exactOutputSingle (SwapRouter02 variant)
	"0x5023b4df", // exactOutput (SwapRouter02 variant)
}

// FunctionSignatureMap maps signatures to function names for quick lookup
var FunctionSignatureMap = map[string]string{
	// SwapRouter (original)
	"0xc04b8d59": "exactInput",
	"0xf28c0498": "exactOutput",
	"0x414bf389": "exactInputSingle",
	"0xdb3e2198": "exactOutputSingle",
	"0xac9650d8": "multicall",
	"0x5ae401dc": "multicall",
	// SwapRouter02 specific
	"0x04e45aaf": "exactInputSingleV2",
	"0xb858183f": "exactInputV2",
	"0x09b81346": "exactOutputSingleV2",
	"0x5023b4df": "exactOutputV2",
}
