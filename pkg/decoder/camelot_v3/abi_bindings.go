// Package camelot_v3 implements decoder for Camelot V3 transactions
package camelot_v3

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Camelot V3 Router ABI for swap functions
const RouterABI = `[{
    "inputs": [
        {
            "internalType": "address",
            "name": "tokenIn",
            "type": "address"
        },
        {
            "internalType": "address",
            "name": "tokenOut",
            "type": "address"
        },
        {
            "internalType": "uint256",
            "name": "amountIn",
            "type": "uint256"
        },
        {
            "internalType": "uint256",
            "name": "minAmountOut",
            "type": "uint256"
        },
        {
            "internalType": "uint128[]",
            "name": "volumes",
            "type": "uint128[]"
        },
        {
            "internalType": "bytes",
            "name": "swapParams",
            "type": "bytes"
        }
    ],
    "name": "swapExactTokensForTokens",
    "outputs": [
        {
            "internalType": "uint256",
            "name": "amountOut",
            "type": "uint256"
        }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
}, {
    "inputs": [
        {
            "internalType": "address",
            "name": "tokenIn",
            "type": "address"
        },
        {
            "internalType": "address",
            "name": "tokenOut",
            "type": "address"
        },
        {
            "internalType": "uint256",
            "name": "amountOut",
            "type": "uint256"
        },
        {
            "internalType": "uint256",
            "name": "maxAmountIn",
            "type": "uint256"
        },
        {
            "internalType": "uint128[]",
            "name": "volumes",
            "type": "uint128[]"
        },
        {
            "internalType": "bytes",
            "name": "swapParams",
            "type": "bytes"
        }
    ],
    "name": "swapTokensForExactTokens",
    "outputs": [
        {
            "internalType": "uint256",
            "name": "amountIn",
            "type": "uint256"
        }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
}, {
    "inputs": [
        {
            "internalType": "address",
            "name": "tokenIn",
            "type": "address"
        },
        {
            "internalType": "address",
            "name": "tokenOut",
            "type": "address"
        },
        {
            "internalType": "uint256",
            "name": "amountIn",
            "type": "uint256"
        },
        {
            "internalType": "uint256",
            "name": "minAmountOut",
            "type": "uint256"
        },
        {
            "internalType": "uint128[]",
            "name": "volumes",
            "type": "uint128[]"
        },
        {
            "internalType": "bytes",
            "name": "swapParams",
            "type": "bytes"
        },
        {
            "internalType": "address",
            "name": "to",
            "type": "address"
        },
        {
            "internalType": "uint256",
            "name": "deadline",
            "type": "uint256"
        }
    ],
    "name": "swapExactTokensForTokensSupportingFeeOnTransferTokens",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
}]`

// ABI instance for Camelot V3 Router
var RouterABIInstance abi.ABI

// Initialize the ABI instance
func init() {
	var err error
	RouterABIInstance, err = abi.JSON(strings.NewReader(RouterABI))
	if err != nil {
		panic(err)
	}
}

// Known Camelot V3 Router addresses on Arbitrum
var KnownRouterAddresses = map[common.Address]string{
	// Arbitrum Camelot V3 Router - actual address should be verified
	common.HexToAddress("0xc873fEcbd354f5A56E00E70921c767647c7A5F2c"): "CamelotV3Router",
}

// Function signatures for Camelot V3 swap operations
var FunctionSignatures = []string{
	"0x12b482a4", // swapExactTokensForTokens
	"0x5b5e066b", // swapTokensForExactTokens (placeholder - actual might differ)
	"0xb614e162", // swapExactTokensForTokensSupportingFeeOnTransferTokens (placeholder - actual might differ)
}
