// Package camelot_v3 implements decoder for Camelot V3 transactions
package camelot_v3

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoder"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// CamelotV3Decoder implements the decoder interface for Camelot V3
type CamelotV3Decoder struct {
	decoder.BaseDecoder
}

// NewCamelotV3Decoder creates a new Camelot V3 decoder
func NewCamelotV3Decoder() *CamelotV3Decoder {
	return &CamelotV3Decoder{
		BaseDecoder: decoder.BaseDecoder{
			ProtocolType: types.CamelotV3Protocol,
			Name:         "CamelotV3",
		},
	}
}

// Matches checks if a transaction should be decoded by this decoder
func (d *CamelotV3Decoder) Matches(tx *ethtypes.Transaction, toAddress string) bool {
	addr := common.HexToAddress(toAddress)

	// Check if the destination address is a known Camelot V3 router
	if _, exists := KnownRouterAddresses[addr]; !exists {
		return false
	}

	// Check if the transaction matches any Camelot V3 function signatures
	return decoder.MatchesSignature(tx, FunctionSignatures)
}

// Decode decodes the transaction data and returns the decoded actions
func (d *CamelotV3Decoder) Decode(tx *ethtypes.Transaction, toAddress string) ([]types.DecodedAction, error) {
	calldata := tx.Data()
	if len(calldata) < 4 {
		return nil, errors.New("calldata too short")
	}

	// Extract function signature (first 4 bytes)
	signature := "0x" + common.Bytes2Hex(calldata[:4])

	// Decode based on the function signature
	switch signature {
	case "0x12b482a4": // swapExactTokensForTokens
		return d.decodeSwapExactTokensForTokens(calldata)
	case "0x5b5e066b": // swapTokensForExactTokens
		return d.decodeSwapTokensForExactTokens(calldata)
	case "0xb614e162": // swapExactTokensForTokensSupportingFeeOnTransferTokens
		return d.decodeSwapExactTokensForTokensSupportingFeeOnTransferTokens(calldata)
	default:
		return nil, fmt.Errorf("unknown Camelot V3 function signature: %s", signature)
	}
}

// Protocol returns the protocol this decoder handles
func (d *CamelotV3Decoder) Protocol() types.ProtocolType {
	return d.ProtocolType
}

// decodeSwapExactTokensForTokens decodes the swapExactTokensForTokens function parameters
func (d *CamelotV3Decoder) decodeSwapExactTokensForTokens(calldata []byte) ([]types.DecodedAction, error) {
	// Parse the ABI-encoded parameters
	args := RouterABIInstance.Methods["swapExactTokensForTokens"]

	// Decode the input parameters
	v, err := args.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack swapExactTokensForTokens parameters: %w", err)
	}

	// Extract parameters
	tokenIn := v[0].(common.Address)
	tokenOut := v[1].(common.Address)
	amountIn := v[2].(*big.Int)
	minAmountOut := v[3].(*big.Int)
	volumes := v[4].([]interface{}) // Volumes array
	// swapParams := v[5].([]byte) // Swap parameters, not needed for basic decoding

	action := types.DecodedAction{
		Type:      types.SwapAction,
		Protocol:  types.CamelotV3Protocol,
		TokenIn:   types.Token{Address: tokenIn},
		TokenOut:  types.Token{Address: tokenOut},
		AmountIn:  amountIn,
		AmountOut: minAmountOut, // This is technically the minimum amount out, not the actual
		Params: map[string]interface{}{
			"function":     "swapExactTokensForTokens",
			"minAmountOut": minAmountOut,
			"volumes":      volumes,
			// "swapParams": swapParams, // Omitting complex swap params for now
		},
	}

	return []types.DecodedAction{action}, nil
}

// decodeSwapTokensForExactTokens decodes the swapTokensForExactTokens function parameters
func (d *CamelotV3Decoder) decodeSwapTokensForExactTokens(calldata []byte) ([]types.DecodedAction, error) {
	// Parse the ABI-encoded parameters - this is a placeholder structure with real ABI parameters
	args := RouterABIInstance.Methods["swapTokensForExactTokens"]

	// Decode the input parameters
	v, err := args.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack swapTokensForExactTokens parameters: %w", err)
	}

	// Extract parameters
	tokenIn := v[0].(common.Address)
	tokenOut := v[1].(common.Address)
	amountOut := v[2].(*big.Int)
	maxAmountIn := v[3].(*big.Int)
	volumes := v[4].([]interface{}) // Volumes array
	// swapParams := v[5].([]byte) // Swap parameters, not needed for basic decoding

	action := types.DecodedAction{
		Type:      types.SwapAction,
		Protocol:  types.CamelotV3Protocol,
		TokenIn:   types.Token{Address: tokenIn},
		TokenOut:  types.Token{Address: tokenOut},
		AmountIn:  maxAmountIn, // This is technically the maximum amount in, not the actual
		AmountOut: amountOut,
		Params: map[string]interface{}{
			"function":    "swapTokensForExactTokens",
			"maxAmountIn": maxAmountIn,
			"volumes":     volumes,
			// "swapParams": swapParams, // Omitting complex swap params for now
		},
	}

	return []types.DecodedAction{action}, nil
}

// decodeSwapExactTokensForTokensSupportingFeeOnTransferTokens decodes the swapExactTokensForTokensSupportingFeeOnTransferTokens function parameters
func (d *CamelotV3Decoder) decodeSwapExactTokensForTokensSupportingFeeOnTransferTokens(calldata []byte) ([]types.DecodedAction, error) {
	// Parse the ABI-encoded parameters - this is a placeholder structure with real ABI parameters
	args := RouterABIInstance.Methods["swapExactTokensForTokensSupportingFeeOnTransferTokens"]

	// Decode the input parameters
	v, err := args.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack swapExactTokensForTokensSupportingFeeOnTransferTokens parameters: %w", err)
	}

	// Extract parameters
	tokenIn := v[0].(common.Address)
	tokenOut := v[1].(common.Address)
	amountIn := v[2].(*big.Int)
	minAmountOut := v[3].(*big.Int)
	volumes := v[4].([]interface{}) // Volumes array
	// swapParams := v[5].([]byte) // Swap parameters, not needed for basic decoding
	recipient := v[6].(common.Address)
	deadline := v[7].(*big.Int)

	action := types.DecodedAction{
		Type:      types.SwapAction,
		Protocol:  types.CamelotV3Protocol,
		TokenIn:   types.Token{Address: tokenIn},
		TokenOut:  types.Token{Address: tokenOut},
		AmountIn:  amountIn,
		AmountOut: minAmountOut, // This is technically the minimum amount out, not the actual
		Params: map[string]interface{}{
			"function":     "swapExactTokensForTokensSupportingFeeOnTransferTokens",
			"minAmountOut": minAmountOut,
			"volumes":      volumes,
			"recipient":    recipient,
			"deadline":     deadline,
			// "swapParams":  swapParams, // Omitting complex swap params for now
		},
	}

	return []types.DecodedAction{action}, nil
}
