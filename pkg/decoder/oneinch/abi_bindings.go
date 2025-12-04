// Package oneinch implements decoder for 1inch Aggregator transactions
package oneinch

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// 1inch AggregationRouter V5 ABI for swap functions
const AggregatorABI = `[
{
    "inputs": [
        {"internalType": "address", "name": "executor", "type": "address"},
        {
            "components": [
                {"internalType": "address", "name": "srcToken", "type": "address"},
                {"internalType": "address", "name": "dstToken", "type": "address"},
                {"internalType": "address", "name": "srcReceiver", "type": "address"},
                {"internalType": "address", "name": "dstReceiver", "type": "address"},
                {"internalType": "uint256", "name": "amount", "type": "uint256"},
                {"internalType": "uint256", "name": "minReturnAmount", "type": "uint256"},
                {"internalType": "uint256", "name": "flags", "type": "uint256"}
            ],
            "internalType": "struct GenericRouter.SwapDescription",
            "name": "desc",
            "type": "tuple"
        },
        {"internalType": "bytes", "name": "permit", "type": "bytes"},
        {"internalType": "bytes", "name": "data", "type": "bytes"}
    ],
    "name": "swap",
    "outputs": [
        {"internalType": "uint256", "name": "returnAmount", "type": "uint256"},
        {"internalType": "uint256", "name": "spentAmount", "type": "uint256"}
    ],
    "stateMutability": "payable",
    "type": "function"
},
{
    "inputs": [
        {"internalType": "address", "name": "srcToken", "type": "address"},
        {"internalType": "uint256", "name": "amount", "type": "uint256"},
        {"internalType": "uint256", "name": "minReturn", "type": "uint256"},
        {"internalType": "uint256[]", "name": "pools", "type": "uint256[]"}
    ],
    "name": "unoswap",
    "outputs": [{"internalType": "uint256", "name": "returnAmount", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
},
{
    "inputs": [
        {"internalType": "address payable", "name": "recipient", "type": "address"},
        {"internalType": "address", "name": "srcToken", "type": "address"},
        {"internalType": "uint256", "name": "amount", "type": "uint256"},
        {"internalType": "uint256", "name": "minReturn", "type": "uint256"},
        {"internalType": "uint256[]", "name": "pools", "type": "uint256[]"}
    ],
    "name": "unoswapTo",
    "outputs": [{"internalType": "uint256", "name": "returnAmount", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
},
{
    "inputs": [
        {"internalType": "uint256", "name": "amount", "type": "uint256"},
        {"internalType": "uint256", "name": "minReturn", "type": "uint256"},
        {"internalType": "uint256[]", "name": "pools", "type": "uint256[]"}
    ],
    "name": "uniswapV3Swap",
    "outputs": [{"internalType": "uint256", "name": "returnAmount", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
},
{
    "inputs": [
        {"internalType": "address payable", "name": "recipient", "type": "address"},
        {"internalType": "uint256", "name": "amount", "type": "uint256"},
        {"internalType": "uint256", "name": "minReturn", "type": "uint256"},
        {"internalType": "uint256[]", "name": "pools", "type": "uint256[]"}
    ],
    "name": "uniswapV3SwapTo",
    "outputs": [{"internalType": "uint256", "name": "returnAmount", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
}
]`

var (
	aggregatorABIOnce    sync.Once
	aggregatorABI        abi.ABI
	aggregatorABIInitErr error
)

// GetAggregatorABI returns the parsed 1inch Aggregator ABI
func GetAggregatorABI() (abi.ABI, error) {
	aggregatorABIOnce.Do(func() {
		aggregatorABI, aggregatorABIInitErr = abi.JSON(strings.NewReader(AggregatorABI))
	})
	return aggregatorABI, aggregatorABIInitErr
}

// Known 1inch router addresses on Arbitrum
var KnownRouterAddresses = map[common.Address]string{
	common.HexToAddress("0x1111111254EEB25477B68fb85Ed929f73A960582"): "1inchAggregatorV5",
	common.HexToAddress("0x1111111254fb6c44bAC0beD2854e76F90643097d"): "1inchAggregatorV4",
}

// Function signatures for 1inch operations
var FunctionSignatures = []string{
	"0x12aa3caf", // swap
	"0x0502b1c5", // unoswap
	"0x2e95b6c8", // unoswapTo
	"0xe449022e", // uniswapV3Swap
	"0xbc80f1a8", // uniswapV3SwapTo
}

// FunctionSignatureMap maps signatures to function names
var FunctionSignatureMap = map[string]string{
	"0x12aa3caf": "swap",
	"0x0502b1c5": "unoswap",
	"0x2e95b6c8": "unoswapTo",
	"0xe449022e": "uniswapV3Swap",
	"0xbc80f1a8": "uniswapV3SwapTo",
}
