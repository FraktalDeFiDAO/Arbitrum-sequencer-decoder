// Package classifier identifies DEX transactions by address and function signature
package classifier

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// Known DEX addresses on Arbitrum
var KnownDEXAddresses = map[common.Address]string{
	// Uniswap V3 - Original SwapRouter
	common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564"): "UniswapV3SwapRouter",
	// Uniswap V3 - SwapRouter02
	common.HexToAddress("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45"): "UniswapV3SwapRouter02",
	// Uniswap Universal Router (most commonly used on Arbitrum)
	common.HexToAddress("0x5E325eDA8064b456f4781070C0738d849c824258"): "UniswapUniversalRouter",
	// Uniswap Universal Router V2
	common.HexToAddress("0x4C60051384bd2d3C01bfc845Cf5F4b44bcbE9de5"): "UniswapUniversalRouterV2",

	// Camelot V3
	common.HexToAddress("0xc873fEcbd354f5A56E00E70921c767647c7A5F2c"): "CamelotV3Router",
	// Camelot V2 Router
	common.HexToAddress("0xc873fEcbd354f5A56E00E70921c767647c7A5F2c"): "CamelotRouter",

	// SushiSwap
	common.HexToAddress("0x1b02dA8Cb0d097eB8D57A175b88c7D8b47997506"): "SushiSwapRouter",
	common.HexToAddress("0x8A21F6768C1f8075791D08546D6FaF9C2d9e8c58"): "SushiSwapRouteProcessor",

	// GMX
	common.HexToAddress("0xaBBc5F99639c9B6bCb58544ddf04EFA6802F4064"): "GMXRouter",
	common.HexToAddress("0xA906F338CB21815cBc4Bc87ace9e68c87eF8d8F1"): "GMXRouter2",
	common.HexToAddress("0x87a4088Bd721F83b6c2E5102e2FA47022Cb1c831"): "GMXExchangeRouter",

	// 1inch
	common.HexToAddress("0x1111111254EEB25477B68fb85Ed929f73A960582"): "1inchAggregatorV5",
	common.HexToAddress("0x1111111254fb6c44bAC0beD2854e76F90643097d"): "1inchAggregatorV4",

	// Paraswap
	common.HexToAddress("0xDEF171Fe48CF0115B1d80b88dc8eAB59176FEe57"): "ParaswapRouter",

	// 0x Protocol
	common.HexToAddress("0xDef1C0ded9bec7F1a1670819833240f027b25EfF"): "0xExchangeProxy",

	// Curve
	common.HexToAddress("0x4e3b3F2b41De373a8e49eDe8aE456e3f66A12bE4"): "CurveV1Router",
	common.HexToAddress("0xF0d4c12A5768D806021F80a262B4d39d26C58b8D"): "CurveRouter",

	// Balancer V2
	common.HexToAddress("0xBA12222222228d8Ba445958a75a0704d566BF2C8"): "BalancerV2Vault",

	// Ramses
	common.HexToAddress("0xAAA87963EAdFc8a017b7811CE6D3b4E4b4D5Dc11"): "RamsesV2Router",

	// Kyber
	common.HexToAddress("0x6131B5fae19EA4f9D964eAc0408E4408b66337b5"): "KyberSwapRouter",
	common.HexToAddress("0x4f9b7DEDD8865871dF65c5D3593CaCE3b7FA3349"): "KyberElasticRouter",

	// Trader Joe
	common.HexToAddress("0xbeE5c10Cf6E4F68f831E11C1D9E59B43560B3571"): "TraderJoeLBRouter",
	common.HexToAddress("0xb4315e873dBcf96Ffd0acd8EA43f689D8c20fB30"): "TraderJoeRouterV2",

	// Odos
	common.HexToAddress("0xa669e7A0d4b3e4Fa48af2dE86BD4CD7126Be4e13"): "OdosRouterV2",

	// OpenOcean
	common.HexToAddress("0x6352a56caadc4F1E25CD6c75970Fa768A3304e64"): "OpenOceanExchange",

	// WooFi
	common.HexToAddress("0x4c4AF8DBc524681930a27b2F1Af5bcC8062E6fB7"): "WooFiRouter",
}

// Function signatures for DEX operations
var FunctionSignatures = map[string]string{
	// Uniswap V3 SwapRouter
	"0xc04b8d59": "UniswapV3: exactInput",
	"0xf28c0498": "UniswapV3: exactOutput",
	"0x414bf389": "UniswapV3: exactInputSingle",
	"0xdb3e2198": "UniswapV3: exactOutputSingle",
	"0xac9650d8": "UniswapV3: multicall",
	"0x5ae401dc": "UniswapV3: multicall2",
	// SwapRouter02 specific
	"0x04e45aaf": "UniswapV3: exactInputSingle",  // SwapRouter02 variant
	"0xb858183f": "UniswapV3: exactInput",        // SwapRouter02 variant
	"0x09b81346": "UniswapV3: exactOutputSingle", // SwapRouter02 variant
	"0x5023b4df": "UniswapV3: exactOutput",       // SwapRouter02 variant

	// Uniswap Universal Router
	"0x3593564c": "UniversalRouter: execute",
	"0x24856bc3": "UniversalRouter: execute",

	// Uniswap V2 style
	"0x791ac947": "UniswapV2: swapExactTokensForETHSupportingFeeOnTransferTokens",
	"0x5c11d795": "UniswapV2: swapExactTokensForTokensSupportingFeeOnTransferTokens",
	"0x38ed1739": "UniswapV2: swapExactTokensForTokens",
	"0x7ff36ab5": "UniswapV2: swapExactETHForTokens",
	"0x18cbafe5": "UniswapV2: swapExactTokensForETH",
	"0xfb3bdb41": "UniswapV2: swapETHForExactTokens",
	"0x4a25d94a": "UniswapV2: swapTokensForExactETH",
	"0x8803dbee": "UniswapV2: swapTokensForExactTokens",
	"0xb6f9de95": "UniswapV2: swapExactETHForTokensSupportingFeeOnTransferTokens",

	// Camelot
	"0x12b482a4": "Camelot: swapExactTokensForTokens",
	"0xd06ca61f": "Camelot: getAmountsOut",

	// SushiSwap RouteProcessor
	"0x2646478b": "Sushi: processRoute",

	// GMX
	"0xea20b34f": "GMX: multicall",
	"0xa94c6bf0": "GMX: swap",
	"0x02de42eb": "GMX: createOrder",

	// 1inch
	"0x12aa3caf": "1inch: swap",
	"0x0502b1c5": "1inch: unoswap",
	"0x2e95b6c8": "1inch: unoswapTo",
	"0xe449022e": "1inch: uniswapV3Swap",
	"0xbc80f1a8": "1inch: uniswapV3SwapTo",

	// Paraswap
	"0x54e3f31b": "Paraswap: simpleSwap",
	"0x64466805": "Paraswap: multiSwap",
	"0x0b86a4c1": "Paraswap: megaSwap",
	"0x46c67b6d": "Paraswap: simpleBuy",

	// 0x Protocol
	"0xd9627aa4": "0x: sellToUniswap",
	"0xf7fcd384": "0x: sellToLiquidityProvider",
	"0x7a1eb1b9": "0x: transformERC20",
	"0x415565b0": "0x: transformERC20", // Different signature variant

	// Curve
	"0xd85b4e8b": "Curve: exchange",
	"0x3df02124": "Curve: exchange_underlying",
	"0x5b41b908": "Curve: exchange_multiple",

	// Balancer V2
	"0x6317ec1b": "Balancer: swap",
	"0x52bbbe29": "Balancer: batchSwap",
	"0x945bcec9": "Balancer: flashLoan",

	// Ramses
	"0x0b4c7e25": "Ramses: swapExactAmounts",

	// Kyber (0xe21fd0e9 = swap with SwapExecutionParams)
	"0xe21fd0e9": "Kyber: swap",

	// Trader Joe
	"0xe5e31b13": "TraderJoe: swapExactTokensForTokens",
	"0xc1c5fe93": "TraderJoe: swapTokensForExactTokens",

	// Odos
	"0x83bd37f9": "Odos: swapCompact",
	"0x84a7f3dd": "Odos: swapMultiCompact",
	"0x3b635ce4": "Odos: swap",
	"0x7bf2d6d4": "Odos: swapMulti",

	// OpenOcean
	"0x90411a32": "OpenOcean: swap",
	"0xa1251d75": "OpenOcean: callUniswapTo",

	// WooFi
	"0x7dc20382": "WooFi: swap",
}

// IsDEXTransaction checks if a transaction is to a known DEX contract
func IsDEXTransaction(toAddress common.Address) bool {
	_, exists := KnownDEXAddresses[toAddress]
	return exists
}

// GetDEXProtocol returns the protocol name for a DEX address
func GetDEXProtocol(toAddress common.Address) string {
	return KnownDEXAddresses[toAddress]
}

// GetFunctionType returns the function type based on the calldata signature
func GetFunctionType(calldata []byte) string {
	if len(calldata) < 4 {
		return ""
	}

	// Extract the first 4 bytes which represent the function signature
	signature := "0x" + common.Bytes2Hex(calldata[:4])
	return FunctionSignatures[signature]
}

// ClassifyTransaction classifies a transaction based on its destination and calldata
func ClassifyTransaction(toAddress common.Address, calldata []byte) (string, string, bool) {
	// Check if the destination is a known DEX contract
	protocol := GetDEXProtocol(toAddress)
	if protocol == "" {
		return "", "", false
	}

	// Determine the function type based on calldata
	functionType := GetFunctionType(calldata)

	// If we can't identify the specific function, return the protocol with a generic DEX type
	if functionType == "" {
		return protocol, "GenericDEX", true
	}

	return protocol, strings.Split(functionType, ": ")[1], true
}
