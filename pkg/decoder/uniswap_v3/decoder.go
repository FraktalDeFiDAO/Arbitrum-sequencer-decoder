// Package uniswap_v3 implements decoder for Uniswap V3 transactions
package uniswap_v3

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoder"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
)

// UniswapV3Decoder implements the decoder interface for Uniswap V3
type UniswapV3Decoder struct {
	decoder.BaseDecoder
}

// NewUniswapV3Decoder creates a new Uniswap V3 decoder
func NewUniswapV3Decoder() *UniswapV3Decoder {
	return &UniswapV3Decoder{
		BaseDecoder: decoder.BaseDecoder{
			ProtocolType: types.UniswapV3Protocol,
			Name:         "UniswapV3",
		},
	}
}

// Matches checks if a transaction should be decoded by this decoder
func (d *UniswapV3Decoder) Matches(tx *ethtypes.Transaction, toAddress string) bool {
	addr := common.HexToAddress(toAddress)

	// Check if the destination address is a known Uniswap V3 router
	if _, exists := KnownRouterAddresses[addr]; !exists {
		return false
	}

	// Check if the transaction matches any Uniswap V3 function signatures
	return decoder.MatchesSignature(tx, FunctionSignatures)
}

// Decode decodes the transaction data and returns the decoded actions
func (d *UniswapV3Decoder) Decode(tx *ethtypes.Transaction, toAddress string) ([]types.DecodedAction, error) {
	calldata := tx.Data()
	if len(calldata) < 4 {
		return nil, errors.New("calldata too short")
	}

	// Extract function signature (first 4 bytes)
	signature := "0x" + common.Bytes2Hex(calldata[:4])

	// Decode based on the function signature
	switch signature {
	// Original SwapRouter
	case "0xc04b8d59": // exactInput
		return d.decodeExactInput(calldata)
	case "0xf28c0498": // exactOutput
		return d.decodeExactOutput(calldata)
	case "0x414bf389": // exactInputSingle
		return d.decodeExactInputSingle(calldata)
	case "0xdb3e2198": // exactOutputSingle
		return d.decodeExactOutputSingle(calldata)
	case "0xac9650d8", "0x5ae401dc": // multicall variants
		return d.decodeMulticall(calldata)
	// SwapRouter02 (different param structures, no deadline in struct)
	case "0x04e45aaf": // exactInputSingle V2
		return d.decodeExactInputSingleV2(calldata)
	case "0xb858183f": // exactInput V2
		return d.decodeExactInputV2(calldata)
	case "0x09b81346": // exactOutputSingle V2
		return d.decodeExactOutputSingleV2(calldata)
	case "0x5023b4df": // exactOutput V2
		return d.decodeExactOutputV2(calldata)
	default:
		return nil, fmt.Errorf("unknown Uniswap V3 function signature: %s", signature)
	}
}

// Protocol returns the protocol this decoder handles
func (d *UniswapV3Decoder) Protocol() types.ProtocolType {
	return d.ProtocolType
}

// ExactInputParams represents the params struct for exactInput
type ExactInputParams struct {
	Path             []byte
	Recipient        common.Address
	Deadline         *big.Int
	AmountIn         *big.Int
	AmountOutMinimum *big.Int
}

// ExactInputSingleParams represents the params struct for exactInputSingle
type ExactInputSingleParams struct {
	TokenIn           common.Address
	TokenOut          common.Address
	Fee               *big.Int // uint24
	Recipient         common.Address
	Deadline          *big.Int
	AmountIn          *big.Int
	AmountOutMinimum  *big.Int
	SqrtPriceLimitX96 *big.Int // uint160
}

// ExactOutputParams represents the params struct for exactOutput
type ExactOutputParams struct {
	Path            []byte
	Recipient       common.Address
	Deadline        *big.Int
	AmountOut       *big.Int
	AmountInMaximum *big.Int
}

// ExactOutputSingleParams represents the params struct for exactOutputSingle
type ExactOutputSingleParams struct {
	TokenIn           common.Address
	TokenOut          common.Address
	Fee               *big.Int // uint24
	Recipient         common.Address
	Deadline          *big.Int
	AmountOut         *big.Int
	AmountInMaximum   *big.Int
	SqrtPriceLimitX96 *big.Int // uint160
}

// decodeExactInput decodes the exactInput function parameters
func (d *UniswapV3Decoder) decodeExactInput(calldata []byte) ([]types.DecodedAction, error) {
	routerABI, err := GetRouterABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get router ABI: %w", err)
	}

	method, exists := routerABI.Methods["exactInput"]
	if !exists {
		return nil, errors.New("exactInput method not found in ABI")
	}

	// Decode the input parameters
	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack exactInput parameters: %w", err)
	}

	if len(values) == 0 {
		return nil, errors.New("no parameters unpacked from exactInput")
	}

	// The ABI unpacks the tuple as a struct - extract fields
	params, err := extractExactInputParams(values[0])
	if err != nil {
		return nil, fmt.Errorf("failed to extract exactInput params: %w", err)
	}

	// Decode the path to determine tokens
	path, err := DecodePath(params.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to decode path in exactInput: %w", err)
	}

	if len(path) < 2 {
		return nil, errors.New("path must have at least 2 tokens")
	}

	// Create decoded action for each swap in the path
	// For multi-hop swaps, only the first hop has the known amountIn
	// Intermediate amounts are determined at execution time
	var actions []types.DecodedAction
	totalHops := len(path) - 1
	for i := 0; i < totalHops; i++ {
		var amountIn, amountOut *big.Int
		if i == 0 {
			// First hop: use the specified input amount
			amountIn = params.AmountIn
		} else {
			// Intermediate hops: amount is determined by previous swap output
			amountIn = nil // Unknown until execution
		}
		if i == totalHops-1 {
			// Last hop: the minimum output applies
			amountOut = params.AmountOutMinimum
		} else {
			// Intermediate hops: output feeds into next swap
			amountOut = nil // Unknown until execution
		}

		action := types.DecodedAction{
			Type:      types.SwapAction,
			Protocol:  types.UniswapV3Protocol,
			TokenIn:   types.Token{Address: path[i]},
			TokenOut:  types.Token{Address: path[i+1]},
			AmountIn:  amountIn,
			AmountOut: amountOut,
			Params: map[string]interface{}{
				"function":          "exactInput",
				"recipient":         params.Recipient,
				"deadline":          params.Deadline,
				"amountOutMinimum":  params.AmountOutMinimum,
				"pathIndex":         i,
				"totalHops":         totalHops,
				"isIntermediateHop": i > 0 && i < totalHops-1,
			},
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// decodeExactOutput decodes the exactOutput function parameters
func (d *UniswapV3Decoder) decodeExactOutput(calldata []byte) ([]types.DecodedAction, error) {
	routerABI, err := GetRouterABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get router ABI: %w", err)
	}

	method, exists := routerABI.Methods["exactOutput"]
	if !exists {
		return nil, errors.New("exactOutput method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack exactOutput parameters: %w", err)
	}

	if len(values) == 0 {
		return nil, errors.New("no parameters unpacked from exactOutput")
	}

	params, err := extractExactOutputParams(values[0])
	if err != nil {
		return nil, fmt.Errorf("failed to extract exactOutput params: %w", err)
	}

	// Decode the path to determine tokens (path is reversed for exactOutput)
	path, err := DecodePath(params.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to decode path in exactOutput: %w", err)
	}

	if len(path) < 2 {
		return nil, errors.New("path must have at least 2 tokens")
	}

	// For exactOutput, the path is reversed (output token first)
	// Only the last hop (first in reversed order) has the known amountOut
	var actions []types.DecodedAction
	totalHops := len(path) - 1
	for i := 0; i < totalHops; i++ {
		var amountIn, amountOut *big.Int
		if i == 0 {
			// First hop in reversed path: this is the final output
			amountOut = params.AmountOut
		} else {
			// Intermediate hops: output determined at execution
			amountOut = nil
		}
		if i == totalHops-1 {
			// Last hop in reversed path: this is where max input applies
			amountIn = params.AmountInMaximum
		} else {
			// Intermediate hops: input determined by next swap requirement
			amountIn = nil
		}

		action := types.DecodedAction{
			Type:      types.SwapAction,
			Protocol:  types.UniswapV3Protocol,
			TokenIn:   types.Token{Address: path[i+1]}, // Reversed
			TokenOut:  types.Token{Address: path[i]},   // Reversed
			AmountIn:  amountIn,
			AmountOut: amountOut,
			Params: map[string]interface{}{
				"function":          "exactOutput",
				"recipient":         params.Recipient,
				"deadline":          params.Deadline,
				"amountInMaximum":   params.AmountInMaximum,
				"pathIndex":         i,
				"totalHops":         totalHops,
				"isIntermediateHop": i > 0 && i < totalHops-1,
			},
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// decodeExactInputSingle decodes the exactInputSingle function parameters
func (d *UniswapV3Decoder) decodeExactInputSingle(calldata []byte) ([]types.DecodedAction, error) {
	routerABI, err := GetRouterABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get router ABI: %w", err)
	}

	method, exists := routerABI.Methods["exactInputSingle"]
	if !exists {
		return nil, errors.New("exactInputSingle method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack exactInputSingle parameters: %w", err)
	}

	if len(values) == 0 {
		return nil, errors.New("no parameters unpacked from exactInputSingle")
	}

	params, err := extractExactInputSingleParams(values[0])
	if err != nil {
		return nil, fmt.Errorf("failed to extract exactInputSingle params: %w", err)
	}

	action := types.DecodedAction{
		Type:      types.SwapAction,
		Protocol:  types.UniswapV3Protocol,
		TokenIn:   types.Token{Address: params.TokenIn},
		TokenOut:  types.Token{Address: params.TokenOut},
		AmountIn:  params.AmountIn,
		AmountOut: params.AmountOutMinimum,
		Params: map[string]interface{}{
			"function":          "exactInputSingle",
			"fee":               params.Fee,
			"recipient":         params.Recipient,
			"deadline":          params.Deadline,
			"amountOutMinimum":  params.AmountOutMinimum,
			"sqrtPriceLimitX96": params.SqrtPriceLimitX96,
		},
	}

	return []types.DecodedAction{action}, nil
}

// decodeExactOutputSingle decodes the exactOutputSingle function parameters
func (d *UniswapV3Decoder) decodeExactOutputSingle(calldata []byte) ([]types.DecodedAction, error) {
	routerABI, err := GetRouterABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get router ABI: %w", err)
	}

	method, exists := routerABI.Methods["exactOutputSingle"]
	if !exists {
		return nil, errors.New("exactOutputSingle method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack exactOutputSingle parameters: %w", err)
	}

	if len(values) == 0 {
		return nil, errors.New("no parameters unpacked from exactOutputSingle")
	}

	params, err := extractExactOutputSingleParams(values[0])
	if err != nil {
		return nil, fmt.Errorf("failed to extract exactOutputSingle params: %w", err)
	}

	action := types.DecodedAction{
		Type:      types.SwapAction,
		Protocol:  types.UniswapV3Protocol,
		TokenIn:   types.Token{Address: params.TokenIn},
		TokenOut:  types.Token{Address: params.TokenOut},
		AmountIn:  params.AmountInMaximum,
		AmountOut: params.AmountOut,
		Params: map[string]interface{}{
			"function":          "exactOutputSingle",
			"fee":               params.Fee,
			"recipient":         params.Recipient,
			"deadline":          params.Deadline,
			"amountInMaximum":   params.AmountInMaximum,
			"sqrtPriceLimitX96": params.SqrtPriceLimitX96,
		},
	}

	return []types.DecodedAction{action}, nil
}

// decodeMulticall decodes multicall batched transactions
func (d *UniswapV3Decoder) decodeMulticall(calldata []byte) ([]types.DecodedAction, error) {
	routerABI, err := GetRouterABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get router ABI: %w", err)
	}

	signature := "0x" + common.Bytes2Hex(calldata[:4])
	var methodName string
	if signature == "0xac9650d8" {
		methodName = "multicall"
	} else {
		methodName = "multicall0" // The overloaded version with deadline
	}

	// Try both multicall variants
	method, exists := routerABI.Methods[methodName]
	if !exists {
		// Fall back to the other variant
		method, exists = routerABI.Methods["multicall"]
		if !exists {
			return nil, errors.New("multicall method not found in ABI")
		}
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack multicall parameters: %w", err)
	}

	// Extract the bytes[] data array
	var dataArray [][]byte
	for _, v := range values {
		if data, ok := v.([][]byte); ok {
			dataArray = data
			break
		}
	}

	if len(dataArray) == 0 {
		return nil, errors.New("no calls found in multicall")
	}

	// Decode each inner call directly from its calldata
	var allActions []types.DecodedAction
	for i, data := range dataArray {
		if len(data) < 4 {
			continue // Skip invalid calls
		}

		actions, err := d.decodeCalldata(data)
		if err != nil {
			// Log but continue with other calls
			continue
		}

		// Tag actions with multicall index
		for j := range actions {
			if actions[j].Params == nil {
				actions[j].Params = make(map[string]interface{})
			}
			actions[j].Params["multicallIndex"] = i
		}
		allActions = append(allActions, actions...)
	}

	return allActions, nil
}

// decodeCalldata decodes raw calldata without needing a transaction object
func (d *UniswapV3Decoder) decodeCalldata(calldata []byte) ([]types.DecodedAction, error) {
	if len(calldata) < 4 {
		return nil, errors.New("calldata too short")
	}

	// Extract function selector as hex string with "0x" prefix
	selectorHex := "0x" + hex.EncodeToString(calldata[:4])

	// Look up the function name
	funcName, exists := FunctionSignatureMap[selectorHex]
	if !exists {
		return nil, fmt.Errorf("unrecognized function selector in multicall: %s", selectorHex)
	}

	switch funcName {
	// Original SwapRouter functions
	case "exactInputSingle":
		return d.decodeExactInputSingle(calldata)
	case "exactOutputSingle":
		return d.decodeExactOutputSingle(calldata)
	case "exactInput":
		return d.decodeExactInput(calldata)
	case "exactOutput":
		return d.decodeExactOutput(calldata)
	// SwapRouter02 V2 functions (used in multicall)
	case "exactInputSingleV2":
		return d.decodeExactInputSingleV2(calldata)
	case "exactOutputSingleV2":
		return d.decodeExactOutputSingleV2(calldata)
	case "exactInputV2":
		return d.decodeExactInputV2(calldata)
	case "exactOutputV2":
		return d.decodeExactOutputV2(calldata)
	default:
		// Unrecognized function (shouldn't happen since we checked FunctionSignatureMap)
		return nil, fmt.Errorf("unrecognized function: %s", funcName)
	}
}

// extractExactInputParams extracts ExactInputParams from the unpacked ABI value
func extractExactInputParams(v interface{}) (*ExactInputParams, error) {
	// Use reflection to extract fields from the anonymous struct created by go-ethereum
	rv := reflect.ValueOf(v)

	// Handle pointer types
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// Must be a struct
	if rv.Kind() != reflect.Struct {
		// Fallback: try to extract from interface slice (older go-ethereum versions)
		if arr, ok := v.([]interface{}); ok && len(arr) >= 5 {
			path, _ := arr[0].([]byte)
			recipient, _ := arr[1].(common.Address)
			deadline, _ := arr[2].(*big.Int)
			amountIn, _ := arr[3].(*big.Int)
			amountOutMinimum, _ := arr[4].(*big.Int)

			return &ExactInputParams{
				Path:             path,
				Recipient:        recipient,
				Deadline:         deadline,
				AmountIn:         amountIn,
				AmountOutMinimum: amountOutMinimum,
			}, nil
		}
		return nil, fmt.Errorf("unable to extract ExactInputParams from type %T", v)
	}

	// Extract fields using reflection
	params := &ExactInputParams{}

	if f := rv.FieldByName("Path"); f.IsValid() {
		if path, ok := f.Interface().([]byte); ok {
			params.Path = path
		}
	}
	if f := rv.FieldByName("Recipient"); f.IsValid() {
		if addr, ok := f.Interface().(common.Address); ok {
			params.Recipient = addr
		}
	}
	if f := rv.FieldByName("Deadline"); f.IsValid() {
		if deadline, ok := f.Interface().(*big.Int); ok {
			params.Deadline = deadline
		}
	}
	if f := rv.FieldByName("AmountIn"); f.IsValid() {
		if amountIn, ok := f.Interface().(*big.Int); ok {
			params.AmountIn = amountIn
		}
	}
	if f := rv.FieldByName("AmountOutMinimum"); f.IsValid() {
		if amountOutMinimum, ok := f.Interface().(*big.Int); ok {
			params.AmountOutMinimum = amountOutMinimum
		}
	}

	return params, nil
}

// extractExactOutputParams extracts ExactOutputParams from the unpacked ABI value
func extractExactOutputParams(v interface{}) (*ExactOutputParams, error) {
	// Use reflection to extract fields from the anonymous struct created by go-ethereum
	rv := reflect.ValueOf(v)

	// Handle pointer types
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// Must be a struct
	if rv.Kind() != reflect.Struct {
		// Fallback: try to extract from interface slice (older go-ethereum versions)
		if arr, ok := v.([]interface{}); ok && len(arr) >= 5 {
			path, _ := arr[0].([]byte)
			recipient, _ := arr[1].(common.Address)
			deadline, _ := arr[2].(*big.Int)
			amountOut, _ := arr[3].(*big.Int)
			amountInMaximum, _ := arr[4].(*big.Int)

			return &ExactOutputParams{
				Path:            path,
				Recipient:       recipient,
				Deadline:        deadline,
				AmountOut:       amountOut,
				AmountInMaximum: amountInMaximum,
			}, nil
		}
		return nil, fmt.Errorf("unable to extract ExactOutputParams from type %T", v)
	}

	// Extract fields using reflection
	params := &ExactOutputParams{}

	if f := rv.FieldByName("Path"); f.IsValid() {
		if path, ok := f.Interface().([]byte); ok {
			params.Path = path
		}
	}
	if f := rv.FieldByName("Recipient"); f.IsValid() {
		if addr, ok := f.Interface().(common.Address); ok {
			params.Recipient = addr
		}
	}
	if f := rv.FieldByName("Deadline"); f.IsValid() {
		if deadline, ok := f.Interface().(*big.Int); ok {
			params.Deadline = deadline
		}
	}
	if f := rv.FieldByName("AmountOut"); f.IsValid() {
		if amountOut, ok := f.Interface().(*big.Int); ok {
			params.AmountOut = amountOut
		}
	}
	if f := rv.FieldByName("AmountInMaximum"); f.IsValid() {
		if amountInMaximum, ok := f.Interface().(*big.Int); ok {
			params.AmountInMaximum = amountInMaximum
		}
	}

	return params, nil
}

// extractExactInputSingleParams extracts ExactInputSingleParams from the unpacked ABI value
func extractExactInputSingleParams(v interface{}) (*ExactInputSingleParams, error) {
	// Use reflection to extract fields from the anonymous struct created by go-ethereum
	rv := reflect.ValueOf(v)

	// Handle pointer types
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// Must be a struct
	if rv.Kind() != reflect.Struct {
		// Try as interface slice for older go-ethereum versions
		if arr, ok := v.([]interface{}); ok && len(arr) >= 8 {
			tokenIn, _ := arr[0].(common.Address)
			tokenOut, _ := arr[1].(common.Address)
			fee, _ := arr[2].(*big.Int)
			recipient, _ := arr[3].(common.Address)
			deadline, _ := arr[4].(*big.Int)
			amountIn, _ := arr[5].(*big.Int)
			amountOutMinimum, _ := arr[6].(*big.Int)
			sqrtPriceLimitX96, _ := arr[7].(*big.Int)

			return &ExactInputSingleParams{
				TokenIn:           tokenIn,
				TokenOut:          tokenOut,
				Fee:               fee,
				Recipient:         recipient,
				Deadline:          deadline,
				AmountIn:          amountIn,
				AmountOutMinimum:  amountOutMinimum,
				SqrtPriceLimitX96: sqrtPriceLimitX96,
			}, nil
		}
		return nil, fmt.Errorf("unable to extract ExactInputSingleParams from type %T", v)
	}

	// Extract fields using reflection
	params := &ExactInputSingleParams{}

	if f := rv.FieldByName("TokenIn"); f.IsValid() {
		if addr, ok := f.Interface().(common.Address); ok {
			params.TokenIn = addr
		}
	}
	if f := rv.FieldByName("TokenOut"); f.IsValid() {
		if addr, ok := f.Interface().(common.Address); ok {
			params.TokenOut = addr
		}
	}
	if f := rv.FieldByName("Fee"); f.IsValid() {
		if fee, ok := f.Interface().(*big.Int); ok {
			params.Fee = fee
		}
	}
	if f := rv.FieldByName("Recipient"); f.IsValid() {
		if addr, ok := f.Interface().(common.Address); ok {
			params.Recipient = addr
		}
	}
	if f := rv.FieldByName("Deadline"); f.IsValid() {
		if deadline, ok := f.Interface().(*big.Int); ok {
			params.Deadline = deadline
		}
	}
	if f := rv.FieldByName("AmountIn"); f.IsValid() {
		if amountIn, ok := f.Interface().(*big.Int); ok {
			params.AmountIn = amountIn
		}
	}
	if f := rv.FieldByName("AmountOutMinimum"); f.IsValid() {
		if amountOutMinimum, ok := f.Interface().(*big.Int); ok {
			params.AmountOutMinimum = amountOutMinimum
		}
	}
	if f := rv.FieldByName("SqrtPriceLimitX96"); f.IsValid() {
		if sqrtPriceLimitX96, ok := f.Interface().(*big.Int); ok {
			params.SqrtPriceLimitX96 = sqrtPriceLimitX96
		}
	}

	return params, nil
}

// extractExactOutputSingleParams extracts ExactOutputSingleParams from the unpacked ABI value
func extractExactOutputSingleParams(v interface{}) (*ExactOutputSingleParams, error) {
	// Use reflection to extract fields from the anonymous struct created by go-ethereum
	rv := reflect.ValueOf(v)

	// Handle pointer types
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// Must be a struct
	if rv.Kind() != reflect.Struct {
		// Try as interface slice for older go-ethereum versions
		if arr, ok := v.([]interface{}); ok && len(arr) >= 8 {
			tokenIn, _ := arr[0].(common.Address)
			tokenOut, _ := arr[1].(common.Address)
			fee, _ := arr[2].(*big.Int)
			recipient, _ := arr[3].(common.Address)
			deadline, _ := arr[4].(*big.Int)
			amountOut, _ := arr[5].(*big.Int)
			amountInMaximum, _ := arr[6].(*big.Int)
			sqrtPriceLimitX96, _ := arr[7].(*big.Int)

			return &ExactOutputSingleParams{
				TokenIn:           tokenIn,
				TokenOut:          tokenOut,
				Fee:               fee,
				Recipient:         recipient,
				Deadline:          deadline,
				AmountOut:         amountOut,
				AmountInMaximum:   amountInMaximum,
				SqrtPriceLimitX96: sqrtPriceLimitX96,
			}, nil
		}
		return nil, fmt.Errorf("unable to extract ExactOutputSingleParams from type %T", v)
	}

	// Extract fields using reflection
	params := &ExactOutputSingleParams{}

	if f := rv.FieldByName("TokenIn"); f.IsValid() {
		if addr, ok := f.Interface().(common.Address); ok {
			params.TokenIn = addr
		}
	}
	if f := rv.FieldByName("TokenOut"); f.IsValid() {
		if addr, ok := f.Interface().(common.Address); ok {
			params.TokenOut = addr
		}
	}
	if f := rv.FieldByName("Fee"); f.IsValid() {
		if fee, ok := f.Interface().(*big.Int); ok {
			params.Fee = fee
		}
	}
	if f := rv.FieldByName("Recipient"); f.IsValid() {
		if addr, ok := f.Interface().(common.Address); ok {
			params.Recipient = addr
		}
	}
	if f := rv.FieldByName("Deadline"); f.IsValid() {
		if deadline, ok := f.Interface().(*big.Int); ok {
			params.Deadline = deadline
		}
	}
	if f := rv.FieldByName("AmountOut"); f.IsValid() {
		if amountOut, ok := f.Interface().(*big.Int); ok {
			params.AmountOut = amountOut
		}
	}
	if f := rv.FieldByName("AmountInMaximum"); f.IsValid() {
		if amountInMaximum, ok := f.Interface().(*big.Int); ok {
			params.AmountInMaximum = amountInMaximum
		}
	}
	if f := rv.FieldByName("SqrtPriceLimitX96"); f.IsValid() {
		if sqrtPriceLimitX96, ok := f.Interface().(*big.Int); ok {
			params.SqrtPriceLimitX96 = sqrtPriceLimitX96
		}
	}

	return params, nil
}

// SwapRouter02 V2 param structures (no deadline in struct)
type ExactInputSingleParamsV2 struct {
	TokenIn           common.Address
	TokenOut          common.Address
	Fee               *big.Int
	Recipient         common.Address
	AmountIn          *big.Int
	AmountOutMinimum  *big.Int
	SqrtPriceLimitX96 *big.Int
}

type ExactInputParamsV2 struct {
	Path             []byte
	Recipient        common.Address
	AmountIn         *big.Int
	AmountOutMinimum *big.Int
}

type ExactOutputSingleParamsV2 struct {
	TokenIn           common.Address
	TokenOut          common.Address
	Fee               *big.Int
	Recipient         common.Address
	AmountOut         *big.Int
	AmountInMaximum   *big.Int
	SqrtPriceLimitX96 *big.Int
}

type ExactOutputParamsV2 struct {
	Path            []byte
	Recipient       common.Address
	AmountOut       *big.Int
	AmountInMaximum *big.Int
}

// decodeExactInputSingleV2 decodes SwapRouter02 exactInputSingle
func (d *UniswapV3Decoder) decodeExactInputSingleV2(calldata []byte) ([]types.DecodedAction, error) {
	router02ABI, err := GetRouter02ABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get router02 ABI: %w", err)
	}

	method, exists := router02ABI.Methods["exactInputSingle"]
	if !exists {
		return nil, errors.New("exactInputSingle method not found in Router02 ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack exactInputSingleV2 parameters: %w", err)
	}

	if len(values) == 0 {
		return nil, errors.New("no parameters unpacked from exactInputSingleV2")
	}

	params, err := extractExactInputSingleParamsV2(values[0])
	if err != nil {
		return nil, fmt.Errorf("failed to extract exactInputSingleV2 params: %w", err)
	}

	action := types.DecodedAction{
		Type:      types.SwapAction,
		Protocol:  types.UniswapV3Protocol,
		TokenIn:   types.Token{Address: params.TokenIn},
		TokenOut:  types.Token{Address: params.TokenOut},
		AmountIn:  params.AmountIn,
		AmountOut: params.AmountOutMinimum,
		Params: map[string]interface{}{
			"function":          "exactInputSingleV2",
			"fee":               params.Fee,
			"recipient":         params.Recipient,
			"amountOutMinimum":  params.AmountOutMinimum,
			"sqrtPriceLimitX96": params.SqrtPriceLimitX96,
		},
	}

	return []types.DecodedAction{action}, nil
}

// decodeExactInputV2 decodes SwapRouter02 exactInput
func (d *UniswapV3Decoder) decodeExactInputV2(calldata []byte) ([]types.DecodedAction, error) {
	router02ABI, err := GetRouter02ABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get router02 ABI: %w", err)
	}

	method, exists := router02ABI.Methods["exactInput"]
	if !exists {
		return nil, errors.New("exactInput method not found in Router02 ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack exactInputV2 parameters: %w", err)
	}

	if len(values) == 0 {
		return nil, errors.New("no parameters unpacked from exactInputV2")
	}

	params, err := extractExactInputParamsV2(values[0])
	if err != nil {
		return nil, fmt.Errorf("failed to extract exactInputV2 params: %w", err)
	}

	path, err := DecodePath(params.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to decode path in exactInputV2: %w", err)
	}

	if len(path) < 2 {
		return nil, errors.New("path must have at least 2 tokens")
	}

	var actions []types.DecodedAction
	totalHops := len(path) - 1
	for i := 0; i < totalHops; i++ {
		var amountIn, amountOut *big.Int
		if i == 0 {
			amountIn = params.AmountIn
		}
		if i == totalHops-1 {
			amountOut = params.AmountOutMinimum
		}

		action := types.DecodedAction{
			Type:      types.SwapAction,
			Protocol:  types.UniswapV3Protocol,
			TokenIn:   types.Token{Address: path[i]},
			TokenOut:  types.Token{Address: path[i+1]},
			AmountIn:  amountIn,
			AmountOut: amountOut,
			Params: map[string]interface{}{
				"function":         "exactInputV2",
				"recipient":        params.Recipient,
				"amountOutMinimum": params.AmountOutMinimum,
				"pathIndex":        i,
				"totalHops":        totalHops,
			},
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// decodeExactOutputSingleV2 decodes SwapRouter02 exactOutputSingle
func (d *UniswapV3Decoder) decodeExactOutputSingleV2(calldata []byte) ([]types.DecodedAction, error) {
	router02ABI, err := GetRouter02ABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get router02 ABI: %w", err)
	}

	method, exists := router02ABI.Methods["exactOutputSingle"]
	if !exists {
		return nil, errors.New("exactOutputSingle method not found in Router02 ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack exactOutputSingleV2 parameters: %w", err)
	}

	if len(values) == 0 {
		return nil, errors.New("no parameters unpacked from exactOutputSingleV2")
	}

	params, err := extractExactOutputSingleParamsV2(values[0])
	if err != nil {
		return nil, fmt.Errorf("failed to extract exactOutputSingleV2 params: %w", err)
	}

	action := types.DecodedAction{
		Type:      types.SwapAction,
		Protocol:  types.UniswapV3Protocol,
		TokenIn:   types.Token{Address: params.TokenIn},
		TokenOut:  types.Token{Address: params.TokenOut},
		AmountIn:  params.AmountInMaximum,
		AmountOut: params.AmountOut,
		Params: map[string]interface{}{
			"function":          "exactOutputSingleV2",
			"fee":               params.Fee,
			"recipient":         params.Recipient,
			"amountInMaximum":   params.AmountInMaximum,
			"sqrtPriceLimitX96": params.SqrtPriceLimitX96,
		},
	}

	return []types.DecodedAction{action}, nil
}

// decodeExactOutputV2 decodes SwapRouter02 exactOutput
func (d *UniswapV3Decoder) decodeExactOutputV2(calldata []byte) ([]types.DecodedAction, error) {
	router02ABI, err := GetRouter02ABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get router02 ABI: %w", err)
	}

	method, exists := router02ABI.Methods["exactOutput"]
	if !exists {
		return nil, errors.New("exactOutput method not found in Router02 ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack exactOutputV2 parameters: %w", err)
	}

	if len(values) == 0 {
		return nil, errors.New("no parameters unpacked from exactOutputV2")
	}

	params, err := extractExactOutputParamsV2(values[0])
	if err != nil {
		return nil, fmt.Errorf("failed to extract exactOutputV2 params: %w", err)
	}

	path, err := DecodePath(params.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to decode path in exactOutputV2: %w", err)
	}

	if len(path) < 2 {
		return nil, errors.New("path must have at least 2 tokens")
	}

	var actions []types.DecodedAction
	totalHops := len(path) - 1
	for i := 0; i < totalHops; i++ {
		var amountIn, amountOut *big.Int
		if i == 0 {
			amountOut = params.AmountOut
		}
		if i == totalHops-1 {
			amountIn = params.AmountInMaximum
		}

		action := types.DecodedAction{
			Type:      types.SwapAction,
			Protocol:  types.UniswapV3Protocol,
			TokenIn:   types.Token{Address: path[i+1]},
			TokenOut:  types.Token{Address: path[i]},
			AmountIn:  amountIn,
			AmountOut: amountOut,
			Params: map[string]interface{}{
				"function":        "exactOutputV2",
				"recipient":       params.Recipient,
				"amountInMaximum": params.AmountInMaximum,
				"pathIndex":       i,
				"totalHops":       totalHops,
			},
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// extractExactInputSingleParamsV2 extracts V2 params from unpacked ABI value
func extractExactInputSingleParamsV2(v interface{}) (*ExactInputSingleParamsV2, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("unable to extract ExactInputSingleParamsV2 from type %T", v)
	}

	params := &ExactInputSingleParamsV2{}
	if f := rv.FieldByName("TokenIn"); f.IsValid() {
		if addr, ok := f.Interface().(common.Address); ok {
			params.TokenIn = addr
		}
	}
	if f := rv.FieldByName("TokenOut"); f.IsValid() {
		if addr, ok := f.Interface().(common.Address); ok {
			params.TokenOut = addr
		}
	}
	if f := rv.FieldByName("Fee"); f.IsValid() {
		if fee, ok := f.Interface().(*big.Int); ok {
			params.Fee = fee
		}
	}
	if f := rv.FieldByName("Recipient"); f.IsValid() {
		if addr, ok := f.Interface().(common.Address); ok {
			params.Recipient = addr
		}
	}
	if f := rv.FieldByName("AmountIn"); f.IsValid() {
		if amountIn, ok := f.Interface().(*big.Int); ok {
			params.AmountIn = amountIn
		}
	}
	if f := rv.FieldByName("AmountOutMinimum"); f.IsValid() {
		if amountOutMinimum, ok := f.Interface().(*big.Int); ok {
			params.AmountOutMinimum = amountOutMinimum
		}
	}
	if f := rv.FieldByName("SqrtPriceLimitX96"); f.IsValid() {
		if sqrtPriceLimitX96, ok := f.Interface().(*big.Int); ok {
			params.SqrtPriceLimitX96 = sqrtPriceLimitX96
		}
	}
	return params, nil
}

// extractExactInputParamsV2 extracts V2 params from unpacked ABI value
func extractExactInputParamsV2(v interface{}) (*ExactInputParamsV2, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("unable to extract ExactInputParamsV2 from type %T", v)
	}

	params := &ExactInputParamsV2{}
	if f := rv.FieldByName("Path"); f.IsValid() {
		if path, ok := f.Interface().([]byte); ok {
			params.Path = path
		}
	}
	if f := rv.FieldByName("Recipient"); f.IsValid() {
		if addr, ok := f.Interface().(common.Address); ok {
			params.Recipient = addr
		}
	}
	if f := rv.FieldByName("AmountIn"); f.IsValid() {
		if amountIn, ok := f.Interface().(*big.Int); ok {
			params.AmountIn = amountIn
		}
	}
	if f := rv.FieldByName("AmountOutMinimum"); f.IsValid() {
		if amountOutMinimum, ok := f.Interface().(*big.Int); ok {
			params.AmountOutMinimum = amountOutMinimum
		}
	}
	return params, nil
}

// extractExactOutputSingleParamsV2 extracts V2 params from unpacked ABI value
func extractExactOutputSingleParamsV2(v interface{}) (*ExactOutputSingleParamsV2, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("unable to extract ExactOutputSingleParamsV2 from type %T", v)
	}

	params := &ExactOutputSingleParamsV2{}
	if f := rv.FieldByName("TokenIn"); f.IsValid() {
		if addr, ok := f.Interface().(common.Address); ok {
			params.TokenIn = addr
		}
	}
	if f := rv.FieldByName("TokenOut"); f.IsValid() {
		if addr, ok := f.Interface().(common.Address); ok {
			params.TokenOut = addr
		}
	}
	if f := rv.FieldByName("Fee"); f.IsValid() {
		if fee, ok := f.Interface().(*big.Int); ok {
			params.Fee = fee
		}
	}
	if f := rv.FieldByName("Recipient"); f.IsValid() {
		if addr, ok := f.Interface().(common.Address); ok {
			params.Recipient = addr
		}
	}
	if f := rv.FieldByName("AmountOut"); f.IsValid() {
		if amountOut, ok := f.Interface().(*big.Int); ok {
			params.AmountOut = amountOut
		}
	}
	if f := rv.FieldByName("AmountInMaximum"); f.IsValid() {
		if amountInMaximum, ok := f.Interface().(*big.Int); ok {
			params.AmountInMaximum = amountInMaximum
		}
	}
	if f := rv.FieldByName("SqrtPriceLimitX96"); f.IsValid() {
		if sqrtPriceLimitX96, ok := f.Interface().(*big.Int); ok {
			params.SqrtPriceLimitX96 = sqrtPriceLimitX96
		}
	}
	return params, nil
}

// extractExactOutputParamsV2 extracts V2 params from unpacked ABI value
func extractExactOutputParamsV2(v interface{}) (*ExactOutputParamsV2, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("unable to extract ExactOutputParamsV2 from type %T", v)
	}

	params := &ExactOutputParamsV2{}
	if f := rv.FieldByName("Path"); f.IsValid() {
		if path, ok := f.Interface().([]byte); ok {
			params.Path = path
		}
	}
	if f := rv.FieldByName("Recipient"); f.IsValid() {
		if addr, ok := f.Interface().(common.Address); ok {
			params.Recipient = addr
		}
	}
	if f := rv.FieldByName("AmountOut"); f.IsValid() {
		if amountOut, ok := f.Interface().(*big.Int); ok {
			params.AmountOut = amountOut
		}
	}
	if f := rv.FieldByName("AmountInMaximum"); f.IsValid() {
		if amountInMaximum, ok := f.Interface().(*big.Int); ok {
			params.AmountInMaximum = amountInMaximum
		}
	}
	return params, nil
}
