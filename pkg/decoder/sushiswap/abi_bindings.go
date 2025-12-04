// Package sushiswap implements decoder for SushiSwap router transactions
package sushiswap

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// SushiSwap RouteProcessor ABI
const RouteProcessorABI = `[
{
    "inputs": [
        {"internalType": "address", "name": "tokenIn", "type": "address"},
        {"internalType": "uint256", "name": "amountIn", "type": "uint256"},
        {"internalType": "address", "name": "tokenOut", "type": "address"},
        {"internalType": "uint256", "name": "amountOutMin", "type": "uint256"},
        {"internalType": "address", "name": "to", "type": "address"},
        {"internalType": "bytes", "name": "route", "type": "bytes"}
    ],
    "name": "processRoute",
    "outputs": [{"internalType": "uint256", "name": "amountOut", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
}
]`

// SushiSwap V2 Router ABI (Uniswap V2 style)
const RouterV2ABI = `[
{
    "inputs": [
        {"internalType": "uint256", "name": "amountIn", "type": "uint256"},
        {"internalType": "uint256", "name": "amountOutMin", "type": "uint256"},
        {"internalType": "address[]", "name": "path", "type": "address[]"},
        {"internalType": "address", "name": "to", "type": "address"},
        {"internalType": "uint256", "name": "deadline", "type": "uint256"}
    ],
    "name": "swapExactTokensForTokens",
    "outputs": [{"internalType": "uint256[]", "name": "amounts", "type": "uint256[]"}],
    "stateMutability": "nonpayable",
    "type": "function"
},
{
    "inputs": [
        {"internalType": "uint256", "name": "amountOutMin", "type": "uint256"},
        {"internalType": "address[]", "name": "path", "type": "address[]"},
        {"internalType": "address", "name": "to", "type": "address"},
        {"internalType": "uint256", "name": "deadline", "type": "uint256"}
    ],
    "name": "swapExactETHForTokens",
    "outputs": [{"internalType": "uint256[]", "name": "amounts", "type": "uint256[]"}],
    "stateMutability": "payable",
    "type": "function"
},
{
    "inputs": [
        {"internalType": "uint256", "name": "amountIn", "type": "uint256"},
        {"internalType": "uint256", "name": "amountOutMin", "type": "uint256"},
        {"internalType": "address[]", "name": "path", "type": "address[]"},
        {"internalType": "address", "name": "to", "type": "address"},
        {"internalType": "uint256", "name": "deadline", "type": "uint256"}
    ],
    "name": "swapExactTokensForETH",
    "outputs": [{"internalType": "uint256[]", "name": "amounts", "type": "uint256[]"}],
    "stateMutability": "nonpayable",
    "type": "function"
},
{
    "inputs": [
        {"internalType": "uint256", "name": "amountIn", "type": "uint256"},
        {"internalType": "uint256", "name": "amountOutMin", "type": "uint256"},
        {"internalType": "address[]", "name": "path", "type": "address[]"},
        {"internalType": "address", "name": "to", "type": "address"},
        {"internalType": "uint256", "name": "deadline", "type": "uint256"}
    ],
    "name": "swapExactTokensForTokensSupportingFeeOnTransferTokens",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
}
]`

var (
	routeProcessorABIOnce    sync.Once
	routeProcessorABI        abi.ABI
	routeProcessorABIInitErr error

	routerV2ABIOnce    sync.Once
	routerV2ABI        abi.ABI
	routerV2ABIInitErr error
)

// GetRouteProcessorABI returns the parsed RouteProcessor ABI
func GetRouteProcessorABI() (abi.ABI, error) {
	routeProcessorABIOnce.Do(func() {
		routeProcessorABI, routeProcessorABIInitErr = abi.JSON(strings.NewReader(RouteProcessorABI))
	})
	return routeProcessorABI, routeProcessorABIInitErr
}

// GetRouterV2ABI returns the parsed Router V2 ABI
func GetRouterV2ABI() (abi.ABI, error) {
	routerV2ABIOnce.Do(func() {
		routerV2ABI, routerV2ABIInitErr = abi.JSON(strings.NewReader(RouterV2ABI))
	})
	return routerV2ABI, routerV2ABIInitErr
}

// Known SushiSwap addresses on Arbitrum
var KnownRouterAddresses = map[common.Address]string{
	common.HexToAddress("0x1b02dA8Cb0d097eB8D57A175b88c7D8b47997506"): "SushiSwapRouter",
	common.HexToAddress("0x8A21F6768C1f8075791D08546D6FaF9C2d9e8c58"): "SushiSwapRouteProcessor",
	common.HexToAddress("0x544bA588efD839d2692Fc31EA991cD39993c135F"): "SushiSwapRouteProcessor3",
}

// Function signatures for SushiSwap operations
var FunctionSignatures = []string{
	"0x2646478b", // processRoute
	"0x38ed1739", // swapExactTokensForTokens
	"0x7ff36ab5", // swapExactETHForTokens
	"0x18cbafe5", // swapExactTokensForETH
	"0x791ac947", // swapExactTokensForETHSupportingFeeOnTransferTokens
	"0x5c11d795", // swapExactTokensForTokensSupportingFeeOnTransferTokens
	"0xb6f9de95", // swapExactETHForTokensSupportingFeeOnTransferTokens
}

// FunctionSignatureMap maps signatures to function names
var FunctionSignatureMap = map[string]string{
	"0x2646478b": "processRoute",
	"0x38ed1739": "swapExactTokensForTokens",
	"0x7ff36ab5": "swapExactETHForTokens",
	"0x18cbafe5": "swapExactTokensForETH",
	"0x791ac947": "swapExactTokensForETHSupportingFeeOnTransferTokens",
	"0x5c11d795": "swapExactTokensForTokensSupportingFeeOnTransferTokens",
	"0xb6f9de95": "swapExactETHForTokensSupportingFeeOnTransferTokens",
}
