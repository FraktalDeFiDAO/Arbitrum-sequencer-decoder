// Package zerox implements decoder for 0x Exchange Proxy transactions
package zerox

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Known 0x addresses on Arbitrum
var KnownRouterAddresses = map[common.Address]string{
	common.HexToAddress("0xDef1C0ded9bec7F1a1670819833240f027b25EfF"): "0xExchangeProxy",
}

// Function signatures for 0x Exchange Proxy
var FunctionSignatures = [][]byte{
	{0xd9, 0x62, 0x7a, 0xa4}, // sellToUniswap
	{0x41, 0x55, 0x65, 0xb0}, // transformERC20
	{0xf7, 0xfc, 0xd3, 0x84}, // sellToLiquidityProvider
	{0x7a, 0x1e, 0xb1, 0xb9}, // transformERC20 (alternate)
}

// FunctionSignatureMap maps hex signatures to function names
var FunctionSignatureMap = map[string]string{
	"0xd9627aa4": "sellToUniswap",
	"0x415565b0": "transformERC20",
	"0xf7fcd384": "sellToLiquidityProvider",
	"0x7a1eb1b9": "transformERC20",
}

// 0x Exchange Proxy ABI
const ExchangeProxyABI = `[
{
	"inputs": [
		{"internalType": "address[]", "name": "tokens", "type": "address[]"},
		{"internalType": "uint256", "name": "sellAmount", "type": "uint256"},
		{"internalType": "uint256", "name": "minBuyAmount", "type": "uint256"},
		{"internalType": "bool", "name": "isSushi", "type": "bool"}
	],
	"name": "sellToUniswap",
	"outputs": [{"internalType": "uint256", "name": "buyAmount", "type": "uint256"}],
	"stateMutability": "payable",
	"type": "function"
},
{
	"inputs": [
		{"internalType": "address", "name": "inputToken", "type": "address"},
		{"internalType": "address", "name": "outputToken", "type": "address"},
		{"internalType": "uint256", "name": "inputTokenAmount", "type": "uint256"},
		{"internalType": "uint256", "name": "minOutputTokenAmount", "type": "uint256"},
		{"components": [
			{"internalType": "uint32", "name": "deploymentNonce", "type": "uint32"},
			{"internalType": "bytes", "name": "data", "type": "bytes"}
		], "internalType": "struct ITransformERC20Feature.Transformation[]", "name": "transformations", "type": "tuple[]"}
	],
	"name": "transformERC20",
	"outputs": [{"internalType": "uint256", "name": "outputTokenAmount", "type": "uint256"}],
	"stateMutability": "payable",
	"type": "function"
},
{
	"inputs": [
		{"internalType": "address", "name": "inputToken", "type": "address"},
		{"internalType": "address", "name": "outputToken", "type": "address"},
		{"internalType": "address", "name": "provider", "type": "address"},
		{"internalType": "address", "name": "recipient", "type": "address"},
		{"internalType": "uint256", "name": "sellAmount", "type": "uint256"},
		{"internalType": "uint256", "name": "minBuyAmount", "type": "uint256"},
		{"internalType": "bytes", "name": "auxiliaryData", "type": "bytes"}
	],
	"name": "sellToLiquidityProvider",
	"outputs": [{"internalType": "uint256", "name": "boughtAmount", "type": "uint256"}],
	"stateMutability": "payable",
	"type": "function"
}
]`

var (
	exchangeProxyABI        abi.ABI
	exchangeProxyABIOnce    sync.Once
	exchangeProxyABIInitErr error
)

// GetExchangeProxyABI returns the parsed 0x Exchange Proxy ABI
func GetExchangeProxyABI() (abi.ABI, error) {
	exchangeProxyABIOnce.Do(func() {
		exchangeProxyABI, exchangeProxyABIInitErr = abi.JSON(strings.NewReader(ExchangeProxyABI))
	})
	return exchangeProxyABI, exchangeProxyABIInitErr
}
