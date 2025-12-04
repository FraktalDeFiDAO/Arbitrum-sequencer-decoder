// Package odos implements decoder for Odos router transactions
package odos

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Odos Router V2 ABI for swap functions
const RouterABI = `[
{
    "inputs": [
        {
            "components": [
                {"internalType": "address", "name": "inputToken", "type": "address"},
                {"internalType": "uint256", "name": "inputAmount", "type": "uint256"},
                {"internalType": "address", "name": "inputReceiver", "type": "address"},
                {"internalType": "address", "name": "outputToken", "type": "address"},
                {"internalType": "uint256", "name": "outputQuote", "type": "uint256"},
                {"internalType": "uint256", "name": "outputMin", "type": "uint256"},
                {"internalType": "address", "name": "outputReceiver", "type": "address"}
            ],
            "internalType": "struct OdosRouterV2.swapTokenInfo",
            "name": "tokenInfo",
            "type": "tuple"
        },
        {"internalType": "bytes", "name": "pathDefinition", "type": "bytes"},
        {"internalType": "address", "name": "executor", "type": "address"},
        {"internalType": "uint32", "name": "referralCode", "type": "uint32"}
    ],
    "name": "swap",
    "outputs": [{"internalType": "uint256", "name": "amountOut", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
},
{
    "inputs": [
        {
            "components": [
                {"internalType": "address", "name": "tokenAddress", "type": "address"},
                {"internalType": "uint256", "name": "amountIn", "type": "uint256"},
                {"internalType": "address", "name": "receiver", "type": "address"}
            ],
            "internalType": "struct OdosRouterV2.inputTokenInfo[]",
            "name": "inputs",
            "type": "tuple[]"
        },
        {
            "components": [
                {"internalType": "address", "name": "tokenAddress", "type": "address"},
                {"internalType": "uint256", "name": "relativeValue", "type": "uint256"},
                {"internalType": "address", "name": "receiver", "type": "address"}
            ],
            "internalType": "struct OdosRouterV2.outputTokenInfo[]",
            "name": "outputs",
            "type": "tuple[]"
        },
        {"internalType": "uint256", "name": "valueOutMin", "type": "uint256"},
        {"internalType": "bytes", "name": "pathDefinition", "type": "bytes"},
        {"internalType": "address", "name": "executor", "type": "address"},
        {"internalType": "uint32", "name": "referralCode", "type": "uint32"}
    ],
    "name": "swapMulti",
    "outputs": [{"internalType": "uint256[]", "name": "amountsOut", "type": "uint256[]"}],
    "stateMutability": "payable",
    "type": "function"
},
{
    "inputs": [
        {"internalType": "address", "name": "tokenIn", "type": "address"},
        {"internalType": "uint256", "name": "amountIn", "type": "uint256"},
        {"internalType": "address", "name": "tokenOut", "type": "address"},
        {"internalType": "uint256", "name": "amountOutMin", "type": "uint256"},
        {"internalType": "address", "name": "to", "type": "address"},
        {"internalType": "bytes", "name": "pathDefinition", "type": "bytes"}
    ],
    "name": "swapRouterTokensForTokens",
    "outputs": [{"internalType": "uint256", "name": "amountOut", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
}
]`

var (
	routerABIOnce    sync.Once
	routerABI        abi.ABI
	routerABIInitErr error
)

// GetRouterABI returns the parsed Odos Router ABI
func GetRouterABI() (abi.ABI, error) {
	routerABIOnce.Do(func() {
		routerABI, routerABIInitErr = abi.JSON(strings.NewReader(RouterABI))
	})
	return routerABI, routerABIInitErr
}

// Known Odos addresses on Arbitrum
var KnownRouterAddresses = map[common.Address]string{
	common.HexToAddress("0xa669e7A0d4b3e4Fa48af2dE86BD4CD7126Be4e13"): "OdosRouterV2",
}

// Function signatures for Odos operations
var FunctionSignatures = []string{
	"0x83bd37f9", // swapCompact (compact encoding, used on L2s)
	"0x84a7f3dd", // swapMultiCompact (compact encoding)
	"0x3b635ce4", // swap (standard ABI)
	"0x7bf2d6d4", // swapMulti (standard ABI)
}

// FunctionSignatureMap maps signatures to function names
var FunctionSignatureMap = map[string]string{
	"0x83bd37f9": "swapCompact",
	"0x84a7f3dd": "swapMultiCompact",
	"0x3b635ce4": "swap",
	"0x7bf2d6d4": "swapMulti",
}
