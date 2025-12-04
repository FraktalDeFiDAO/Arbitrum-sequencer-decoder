// Package sushiswap implements decoder for SushiSwap router transactions
package sushiswap

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoder"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
)

// WETH address on Arbitrum
var WETHAddress = common.HexToAddress("0x82aF49447D8a07e3bd95BD0d56f35241523fBab1")

// SushiSwapDecoder implements the decoder interface for SushiSwap
type SushiSwapDecoder struct {
	decoder.BaseDecoder
}

// NewSushiSwapDecoder creates a new SushiSwap decoder
func NewSushiSwapDecoder() *SushiSwapDecoder {
	return &SushiSwapDecoder{
		BaseDecoder: decoder.BaseDecoder{
			ProtocolType: types.SushiSwapProtocol,
			Name:         "SushiSwap",
		},
	}
}

// Matches checks if a transaction should be decoded by this decoder
func (d *SushiSwapDecoder) Matches(tx *ethtypes.Transaction, toAddress string) bool {
	to := common.HexToAddress(toAddress)
	if _, exists := KnownRouterAddresses[to]; !exists {
		return false
	}
	return decoder.MatchesSignature(tx, FunctionSignatures)
}

// Protocol returns the protocol type
func (d *SushiSwapDecoder) Protocol() types.ProtocolType {
	return types.SushiSwapProtocol
}

// Decode decodes the transaction and returns the decoded actions
func (d *SushiSwapDecoder) Decode(tx *ethtypes.Transaction, toAddress string) ([]types.DecodedAction, error) {
	calldata := tx.Data()
	if len(calldata) < 4 {
		return nil, errors.New("calldata too short")
	}

	selectorHex := "0x" + hex.EncodeToString(calldata[:4])
	funcName, exists := FunctionSignatureMap[selectorHex]
	if !exists {
		return nil, fmt.Errorf("unknown function selector: %s", selectorHex)
	}

	switch funcName {
	case "processRoute":
		return d.decodeProcessRoute(calldata)
	case "swapExactTokensForTokens":
		return d.decodeSwapExactTokensForTokens(calldata)
	case "swapExactETHForTokens":
		return d.decodeSwapExactETHForTokens(calldata, tx.Value())
	case "swapExactTokensForETH":
		return d.decodeSwapExactTokensForETH(calldata)
	case "swapExactTokensForTokensSupportingFeeOnTransferTokens":
		return d.decodeSwapExactTokensForTokensSupportingFeeOnTransferTokens(calldata)
	case "swapExactTokensForETHSupportingFeeOnTransferTokens":
		return d.decodeSwapExactTokensForETHSupportingFeeOnTransferTokens(calldata)
	case "swapExactETHForTokensSupportingFeeOnTransferTokens":
		return d.decodeSwapExactETHForTokensSupportingFeeOnTransferTokens(calldata, tx.Value())
	default:
		return nil, fmt.Errorf("unimplemented function: %s", funcName)
	}
}

func (d *SushiSwapDecoder) decodeProcessRoute(calldata []byte) ([]types.DecodedAction, error) {
	routeProcessorABI, err := GetRouteProcessorABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	method, exists := routeProcessorABI.Methods["processRoute"]
	if !exists {
		return nil, errors.New("processRoute method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack processRoute parameters: %w", err)
	}

	if len(values) < 5 {
		return nil, errors.New("insufficient parameters for processRoute")
	}

	tokenIn, _ := values[0].(common.Address)
	amountIn, _ := values[1].(*big.Int)
	tokenOut, _ := values[2].(common.Address)
	amountOutMin, _ := values[3].(*big.Int)
	to, _ := values[4].(common.Address)

	action := types.DecodedAction{
		Type:     types.SwapAction,
		Protocol: types.SushiSwapProtocol,
		TokenIn: types.Token{
			Address: tokenIn,
		},
		TokenOut: types.Token{
			Address: tokenOut,
		},
		AmountIn:  amountIn,
		AmountOut: amountOutMin,
		Params: map[string]interface{}{
			"function":  "processRoute",
			"recipient": to.Hex(),
		},
	}

	return []types.DecodedAction{action}, nil
}

func (d *SushiSwapDecoder) decodeSwapExactTokensForTokens(calldata []byte) ([]types.DecodedAction, error) {
	routerABI, err := GetRouterV2ABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	method, exists := routerABI.Methods["swapExactTokensForTokens"]
	if !exists {
		return nil, errors.New("swapExactTokensForTokens method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack swapExactTokensForTokens parameters: %w", err)
	}

	if len(values) < 4 {
		return nil, errors.New("insufficient parameters for swapExactTokensForTokens")
	}

	amountIn, _ := values[0].(*big.Int)
	amountOutMin, _ := values[1].(*big.Int)
	path, _ := values[2].([]common.Address)

	if len(path) < 2 {
		return nil, errors.New("path must have at least 2 tokens")
	}

	action := types.DecodedAction{
		Type:     types.SwapAction,
		Protocol: types.SushiSwapProtocol,
		TokenIn: types.Token{
			Address: path[0],
		},
		TokenOut: types.Token{
			Address: path[len(path)-1],
		},
		AmountIn:  amountIn,
		AmountOut: amountOutMin,
		Params: map[string]interface{}{
			"function": "swapExactTokensForTokens",
			"pathLen":  len(path),
		},
	}

	// Add path tokens
	if len(path) > 2 {
		action.Path = make([]types.Token, len(path))
		for i, addr := range path {
			action.Path[i] = types.Token{Address: addr}
		}
	}

	return []types.DecodedAction{action}, nil
}

func (d *SushiSwapDecoder) decodeSwapExactETHForTokens(calldata []byte, value *big.Int) ([]types.DecodedAction, error) {
	routerABI, err := GetRouterV2ABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	method, exists := routerABI.Methods["swapExactETHForTokens"]
	if !exists {
		return nil, errors.New("swapExactETHForTokens method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack swapExactETHForTokens parameters: %w", err)
	}

	if len(values) < 2 {
		return nil, errors.New("insufficient parameters for swapExactETHForTokens")
	}

	amountOutMin, _ := values[0].(*big.Int)
	path, _ := values[1].([]common.Address)

	if len(path) < 2 {
		return nil, errors.New("path must have at least 2 tokens")
	}

	action := types.DecodedAction{
		Type:     types.SwapAction,
		Protocol: types.SushiSwapProtocol,
		TokenIn: types.Token{
			Address: WETHAddress, // ETH -> WETH
		},
		TokenOut: types.Token{
			Address: path[len(path)-1],
		},
		AmountIn:  value,
		AmountOut: amountOutMin,
		Params: map[string]interface{}{
			"function": "swapExactETHForTokens",
			"pathLen":  len(path),
		},
	}

	return []types.DecodedAction{action}, nil
}

func (d *SushiSwapDecoder) decodeSwapExactTokensForETH(calldata []byte) ([]types.DecodedAction, error) {
	routerABI, err := GetRouterV2ABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	method, exists := routerABI.Methods["swapExactTokensForETH"]
	if !exists {
		return nil, errors.New("swapExactTokensForETH method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack swapExactTokensForETH parameters: %w", err)
	}

	if len(values) < 3 {
		return nil, errors.New("insufficient parameters for swapExactTokensForETH")
	}

	amountIn, _ := values[0].(*big.Int)
	amountOutMin, _ := values[1].(*big.Int)
	path, _ := values[2].([]common.Address)

	if len(path) < 2 {
		return nil, errors.New("path must have at least 2 tokens")
	}

	action := types.DecodedAction{
		Type:     types.SwapAction,
		Protocol: types.SushiSwapProtocol,
		TokenIn: types.Token{
			Address: path[0],
		},
		TokenOut: types.Token{
			Address: WETHAddress, // -> ETH (WETH)
		},
		AmountIn:  amountIn,
		AmountOut: amountOutMin,
		Params: map[string]interface{}{
			"function": "swapExactTokensForETH",
			"pathLen":  len(path),
		},
	}

	return []types.DecodedAction{action}, nil
}

func (d *SushiSwapDecoder) decodeSwapExactTokensForTokensSupportingFeeOnTransferTokens(calldata []byte) ([]types.DecodedAction, error) {
	routerABI, err := GetRouterV2ABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	method, exists := routerABI.Methods["swapExactTokensForTokensSupportingFeeOnTransferTokens"]
	if !exists {
		return nil, errors.New("swapExactTokensForTokensSupportingFeeOnTransferTokens method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack parameters: %w", err)
	}

	if len(values) < 4 {
		return nil, errors.New("insufficient parameters")
	}

	amountIn, _ := values[0].(*big.Int)
	amountOutMin, _ := values[1].(*big.Int)
	path, _ := values[2].([]common.Address)

	if len(path) < 2 {
		return nil, errors.New("path must have at least 2 tokens")
	}

	action := types.DecodedAction{
		Type:     types.SwapAction,
		Protocol: types.SushiSwapProtocol,
		TokenIn: types.Token{
			Address: path[0],
		},
		TokenOut: types.Token{
			Address: path[len(path)-1],
		},
		AmountIn:  amountIn,
		AmountOut: amountOutMin,
		Params: map[string]interface{}{
			"function":           "swapExactTokensForTokensSupportingFeeOnTransferTokens",
			"feeOnTransferToken": true,
			"pathLen":            len(path),
		},
	}

	return []types.DecodedAction{action}, nil
}

func (d *SushiSwapDecoder) decodeSwapExactTokensForETHSupportingFeeOnTransferTokens(calldata []byte) ([]types.DecodedAction, error) {
	routerABI, err := GetRouterV2ABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	// Uses same ABI as swapExactTokensForETH
	method, exists := routerABI.Methods["swapExactTokensForETH"]
	if !exists {
		return nil, errors.New("swapExactTokensForETH method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack parameters: %w", err)
	}

	if len(values) < 3 {
		return nil, errors.New("insufficient parameters")
	}

	amountIn, _ := values[0].(*big.Int)
	amountOutMin, _ := values[1].(*big.Int)
	path, _ := values[2].([]common.Address)

	if len(path) < 2 {
		return nil, errors.New("path must have at least 2 tokens")
	}

	action := types.DecodedAction{
		Type:     types.SwapAction,
		Protocol: types.SushiSwapProtocol,
		TokenIn: types.Token{
			Address: path[0],
		},
		TokenOut: types.Token{
			Address: WETHAddress, // -> ETH (WETH)
		},
		AmountIn:  amountIn,
		AmountOut: amountOutMin,
		Params: map[string]interface{}{
			"function":           "swapExactTokensForETHSupportingFeeOnTransferTokens",
			"feeOnTransferToken": true,
			"pathLen":            len(path),
		},
	}

	return []types.DecodedAction{action}, nil
}

func (d *SushiSwapDecoder) decodeSwapExactETHForTokensSupportingFeeOnTransferTokens(calldata []byte, value *big.Int) ([]types.DecodedAction, error) {
	routerABI, err := GetRouterV2ABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	// Uses same ABI as swapExactETHForTokens
	method, exists := routerABI.Methods["swapExactETHForTokens"]
	if !exists {
		return nil, errors.New("swapExactETHForTokens method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack parameters: %w", err)
	}

	if len(values) < 2 {
		return nil, errors.New("insufficient parameters")
	}

	amountOutMin, _ := values[0].(*big.Int)
	path, _ := values[1].([]common.Address)

	if len(path) < 2 {
		return nil, errors.New("path must have at least 2 tokens")
	}

	action := types.DecodedAction{
		Type:     types.SwapAction,
		Protocol: types.SushiSwapProtocol,
		TokenIn: types.Token{
			Address: WETHAddress, // ETH -> WETH
		},
		TokenOut: types.Token{
			Address: path[len(path)-1],
		},
		AmountIn:  value,
		AmountOut: amountOutMin,
		Params: map[string]interface{}{
			"function":           "swapExactETHForTokensSupportingFeeOnTransferTokens",
			"feeOnTransferToken": true,
			"pathLen":            len(path),
		},
	}

	return []types.DecodedAction{action}, nil
}

// IsRouterAddress checks if the address is a known SushiSwap router
func IsRouterAddress(addr common.Address) bool {
	addrLower := strings.ToLower(addr.Hex())
	for knownAddr := range KnownRouterAddresses {
		if strings.ToLower(knownAddr.Hex()) == addrLower {
			return true
		}
	}
	return false
}
