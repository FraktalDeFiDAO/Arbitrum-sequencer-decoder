// Package universal_router implements decoder for Uniswap Universal Router transactions
package universal_router

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Known Universal Router addresses on Arbitrum
var KnownRouterAddresses = map[common.Address]string{
	common.HexToAddress("0x5E325eDA8064b456f4781070C0738d849c824258"): "UniswapUniversalRouter",
	common.HexToAddress("0x4C60051384bd2d3C01bfc845Cf5F4b44bcbE9de5"): "UniswapUniversalRouterV2",
}

// Function signatures for Universal Router
var FunctionSignatures = [][]byte{
	{0x35, 0x93, 0x56, 0x4c}, // execute(bytes,bytes[],uint256) - 0x3593564c
	{0x24, 0x85, 0x6b, 0xc3}, // execute(bytes,bytes[]) - 0x24856bc3 (no deadline)
}

// FunctionSignatureMap maps hex signatures to function names
var FunctionSignatureMap = map[string]string{
	"0x3593564c": "execute",
	"0x24856bc3": "executeNoDeadline",
}

// Command types for Universal Router
const (
	// V3 swap commands
	V3_SWAP_EXACT_IN  = 0x00
	V3_SWAP_EXACT_OUT = 0x01

	// V2 swap commands
	V2_SWAP_EXACT_IN  = 0x08
	V2_SWAP_EXACT_OUT = 0x09

	// Other common commands (we skip these but don't error)
	PERMIT2_PERMIT         = 0x0a
	WRAP_ETH               = 0x0b
	UNWRAP_WETH            = 0x0c
	SWEEP                  = 0x04
	TRANSFER               = 0x05
	PAY_PORTION            = 0x06
	PERMIT2_TRANSFER_FROM  = 0x02
	PERMIT2_PERMIT_BATCH   = 0x03
	PERMIT2_TRANSFER_BATCH = 0x07
)

// Universal Router ABI for execute functions
const RouterABI = `[
{
	"inputs": [
		{"internalType": "bytes", "name": "commands", "type": "bytes"},
		{"internalType": "bytes[]", "name": "inputs", "type": "bytes[]"},
		{"internalType": "uint256", "name": "deadline", "type": "uint256"}
	],
	"name": "execute",
	"outputs": [],
	"stateMutability": "payable",
	"type": "function"
},
{
	"inputs": [
		{"internalType": "bytes", "name": "commands", "type": "bytes"},
		{"internalType": "bytes[]", "name": "inputs", "type": "bytes[]"}
	],
	"name": "execute",
	"outputs": [],
	"stateMutability": "payable",
	"type": "function"
}
]`

var (
	routerABI        abi.ABI
	routerABIOnce    sync.Once
	routerABIInitErr error
)

// GetRouterABI returns the parsed Universal Router ABI
func GetRouterABI() (abi.ABI, error) {
	routerABIOnce.Do(func() {
		routerABI, routerABIInitErr = abi.JSON(strings.NewReader(RouterABI))
	})
	return routerABI, routerABIInitErr
}
