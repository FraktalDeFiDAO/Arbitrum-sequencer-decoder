// Package openocean implements decoder for OpenOcean exchange transactions
package openocean

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// OpenOcean Exchange ABI for swap functions
const ExchangeABI = `[
{
    "inputs": [
        {"internalType": "address", "name": "caller", "type": "address"},
        {
            "components": [
                {"internalType": "address", "name": "srcToken", "type": "address"},
                {"internalType": "address", "name": "dstToken", "type": "address"},
                {"internalType": "address", "name": "srcReceiver", "type": "address"},
                {"internalType": "address", "name": "dstReceiver", "type": "address"},
                {"internalType": "uint256", "name": "amount", "type": "uint256"},
                {"internalType": "uint256", "name": "minReturnAmount", "type": "uint256"},
                {"internalType": "uint256", "name": "guaranteedAmount", "type": "uint256"},
                {"internalType": "uint256", "name": "flags", "type": "uint256"},
                {"internalType": "address", "name": "referrer", "type": "address"},
                {"internalType": "bytes", "name": "permit", "type": "bytes"}
            ],
            "internalType": "struct OpenOceanExchange.SwapDescription",
            "name": "desc",
            "type": "tuple"
        },
        {
            "components": [
                {"internalType": "uint256", "name": "target", "type": "uint256"},
                {"internalType": "uint256", "name": "gasLimit", "type": "uint256"},
                {"internalType": "uint256", "name": "value", "type": "uint256"},
                {"internalType": "bytes", "name": "data", "type": "bytes"}
            ],
            "internalType": "struct OpenOceanExchange.CallDescription[]",
            "name": "calls",
            "type": "tuple[]"
        }
    ],
    "name": "swap",
    "outputs": [{"internalType": "uint256", "name": "returnAmount", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
},
{
    "inputs": [
        {"internalType": "contract IOpenOceanCaller", "name": "caller", "type": "address"},
        {"internalType": "address", "name": "srcToken", "type": "address"},
        {"internalType": "address", "name": "dstToken", "type": "address"},
        {"internalType": "uint256", "name": "amount", "type": "uint256"},
        {"internalType": "uint256", "name": "minReturn", "type": "uint256"},
        {"internalType": "uint256[]", "name": "pools", "type": "uint256[]"}
    ],
    "name": "callUniswapTo",
    "outputs": [{"internalType": "uint256", "name": "returnAmount", "type": "uint256"}],
    "stateMutability": "payable",
    "type": "function"
}
]`

var (
	exchangeABIOnce    sync.Once
	exchangeABI        abi.ABI
	exchangeABIInitErr error
)

// GetExchangeABI returns the parsed OpenOcean Exchange ABI
func GetExchangeABI() (abi.ABI, error) {
	exchangeABIOnce.Do(func() {
		exchangeABI, exchangeABIInitErr = abi.JSON(strings.NewReader(ExchangeABI))
	})
	return exchangeABI, exchangeABIInitErr
}

// Known OpenOcean addresses on Arbitrum
var KnownRouterAddresses = map[common.Address]string{
	common.HexToAddress("0x6352a56caadc4F1E25CD6c75970Fa768A3304e64"): "OpenOceanExchange",
}

// Function signatures for OpenOcean operations
var FunctionSignatures = []string{
	"0x90411a32", // swap
	"0xa1251d75", // callUniswapTo
}

// FunctionSignatureMap maps signatures to function names
var FunctionSignatureMap = map[string]string{
	"0x90411a32": "swap",
	"0xa1251d75": "callUniswapTo",
}
