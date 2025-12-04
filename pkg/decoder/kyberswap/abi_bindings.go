// Package kyberswap implements decoder for KyberSwap router transactions
package kyberswap

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// KyberSwap Meta Aggregation Router V2 ABI
// Based on the verified contract at 0x6131B5fae19EA4f9D964eAc0408E4408b66337b5
// 0xe21fd0e9 = swap((address,address,bytes,(address,address,address[],uint256[],address[],uint256[],address,uint256,uint256,uint256,bytes),bytes))
const RouterABI = `[
{
    "inputs": [
        {
            "components": [
                {"internalType": "address", "name": "callTarget", "type": "address"},
                {"internalType": "address", "name": "approveTarget", "type": "address"},
                {"internalType": "bytes", "name": "targetData", "type": "bytes"},
                {
                    "components": [
                        {"internalType": "address", "name": "srcToken", "type": "address"},
                        {"internalType": "address", "name": "dstToken", "type": "address"},
                        {"internalType": "address[]", "name": "srcReceivers", "type": "address[]"},
                        {"internalType": "uint256[]", "name": "srcAmounts", "type": "uint256[]"},
                        {"internalType": "address[]", "name": "feeReceivers", "type": "address[]"},
                        {"internalType": "uint256[]", "name": "feeAmounts", "type": "uint256[]"},
                        {"internalType": "address", "name": "dstReceiver", "type": "address"},
                        {"internalType": "uint256", "name": "amount", "type": "uint256"},
                        {"internalType": "uint256", "name": "minReturnAmount", "type": "uint256"},
                        {"internalType": "uint256", "name": "flags", "type": "uint256"},
                        {"internalType": "bytes", "name": "permit", "type": "bytes"}
                    ],
                    "internalType": "struct MetaAggregationRouterV2.SwapDescription",
                    "name": "desc",
                    "type": "tuple"
                },
                {"internalType": "bytes", "name": "clientData", "type": "bytes"}
            ],
            "internalType": "struct MetaAggregationRouterV2.SwapExecutionParams",
            "name": "execution",
            "type": "tuple"
        }
    ],
    "name": "swap",
    "outputs": [
        {"internalType": "uint256", "name": "returnAmount", "type": "uint256"},
        {"internalType": "uint256", "name": "gasUsed", "type": "uint256"}
    ],
    "stateMutability": "payable",
    "type": "function"
}
]`

var (
	routerABIOnce    sync.Once
	routerABI        abi.ABI
	routerABIInitErr error
)

// GetRouterABI returns the parsed KyberSwap Router ABI
func GetRouterABI() (abi.ABI, error) {
	routerABIOnce.Do(func() {
		routerABI, routerABIInitErr = abi.JSON(strings.NewReader(RouterABI))
	})
	return routerABI, routerABIInitErr
}

// Known KyberSwap addresses on Arbitrum
var KnownRouterAddresses = map[common.Address]string{
	common.HexToAddress("0x6131B5fae19EA4f9D964eAc0408E4408b66337b5"): "KyberSwapMetaAggregatorV2",
	common.HexToAddress("0x4f9b7DEDD8865871dF65c5D3593CaCE3b7FA3349"): "KyberElasticRouter",
}

// Function signatures for KyberSwap operations
// 0xe21fd0e9 = swap((address,address,bytes,(address,address,address[],uint256[],address[],uint256[],address,uint256,uint256,uint256,bytes),bytes))
var FunctionSignatures = []string{
	"0xe21fd0e9", // swap (with SwapExecutionParams)
}

// FunctionSignatureMap maps signatures to function names
var FunctionSignatureMap = map[string]string{
	"0xe21fd0e9": "swap",
}
