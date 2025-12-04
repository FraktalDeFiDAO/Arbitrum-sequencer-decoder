// Package oneinch implements decoder for 1inch Aggregator transactions
package oneinch

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoder"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
)

// OneInchDecoder implements the decoder interface for 1inch Aggregator
type OneInchDecoder struct {
	decoder.BaseDecoder
}

// NewOneInchDecoder creates a new 1inch decoder
func NewOneInchDecoder() *OneInchDecoder {
	return &OneInchDecoder{
		BaseDecoder: decoder.BaseDecoder{
			ProtocolType: types.OneInchProtocol,
			Name:         "1inch Aggregator",
		},
	}
}

// Matches checks if a transaction should be decoded by this decoder
func (d *OneInchDecoder) Matches(tx *ethtypes.Transaction, toAddress string) bool {
	to := common.HexToAddress(toAddress)
	if _, exists := KnownRouterAddresses[to]; !exists {
		return false
	}
	return decoder.MatchesSignature(tx, FunctionSignatures)
}

// Protocol returns the protocol type
func (d *OneInchDecoder) Protocol() types.ProtocolType {
	return types.OneInchProtocol
}

// Decode decodes the transaction and returns the decoded actions
func (d *OneInchDecoder) Decode(tx *ethtypes.Transaction, toAddress string) ([]types.DecodedAction, error) {
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
	case "swap":
		return d.decodeSwap(calldata)
	case "unoswap":
		return d.decodeUnoswap(calldata)
	case "unoswapTo":
		return d.decodeUnoswapTo(calldata)
	case "uniswapV3Swap":
		return d.decodeUniswapV3Swap(calldata)
	case "uniswapV3SwapTo":
		return d.decodeUniswapV3SwapTo(calldata)
	default:
		return nil, fmt.Errorf("unimplemented function: %s", funcName)
	}
}

// SwapDescription represents the swap parameters for 1inch swap
type SwapDescription struct {
	SrcToken        common.Address
	DstToken        common.Address
	SrcReceiver     common.Address
	DstReceiver     common.Address
	Amount          *big.Int
	MinReturnAmount *big.Int
	Flags           *big.Int
}

func (d *OneInchDecoder) decodeSwap(calldata []byte) ([]types.DecodedAction, error) {
	aggregatorABI, err := GetAggregatorABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	method, exists := aggregatorABI.Methods["swap"]
	if !exists {
		return nil, errors.New("swap method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack swap parameters: %w", err)
	}

	if len(values) < 2 {
		return nil, errors.New("insufficient parameters for swap")
	}

	// Extract SwapDescription from values[1]
	desc, err := extractSwapDescription(values[1])
	if err != nil {
		return nil, fmt.Errorf("failed to extract swap description: %w", err)
	}

	action := types.DecodedAction{
		Type:     types.SwapAction,
		Protocol: types.OneInchProtocol,
		TokenIn: types.Token{
			Address: desc.SrcToken,
		},
		TokenOut: types.Token{
			Address: desc.DstToken,
		},
		AmountIn:  desc.Amount,
		AmountOut: desc.MinReturnAmount,
		Params: map[string]interface{}{
			"function":    "swap",
			"srcReceiver": desc.SrcReceiver.Hex(),
			"dstReceiver": desc.DstReceiver.Hex(),
			"flags":       desc.Flags.String(),
		},
	}

	return []types.DecodedAction{action}, nil
}

func (d *OneInchDecoder) decodeUnoswap(calldata []byte) ([]types.DecodedAction, error) {
	aggregatorABI, err := GetAggregatorABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	method, exists := aggregatorABI.Methods["unoswap"]
	if !exists {
		return nil, errors.New("unoswap method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack unoswap parameters: %w", err)
	}

	if len(values) < 4 {
		return nil, errors.New("insufficient parameters for unoswap")
	}

	srcToken, _ := values[0].(common.Address)
	amount, _ := values[1].(*big.Int)
	minReturn, _ := values[2].(*big.Int)

	action := types.DecodedAction{
		Type:     types.SwapAction,
		Protocol: types.OneInchProtocol,
		TokenIn: types.Token{
			Address: srcToken,
		},
		TokenOut:  types.Token{}, // Unknown from calldata alone
		AmountIn:  amount,
		AmountOut: minReturn,
		Params: map[string]interface{}{
			"function": "unoswap",
		},
	}

	return []types.DecodedAction{action}, nil
}

func (d *OneInchDecoder) decodeUnoswapTo(calldata []byte) ([]types.DecodedAction, error) {
	aggregatorABI, err := GetAggregatorABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	method, exists := aggregatorABI.Methods["unoswapTo"]
	if !exists {
		return nil, errors.New("unoswapTo method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack unoswapTo parameters: %w", err)
	}

	if len(values) < 5 {
		return nil, errors.New("insufficient parameters for unoswapTo")
	}

	srcToken, _ := values[1].(common.Address)
	amount, _ := values[2].(*big.Int)
	minReturn, _ := values[3].(*big.Int)

	action := types.DecodedAction{
		Type:     types.SwapAction,
		Protocol: types.OneInchProtocol,
		TokenIn: types.Token{
			Address: srcToken,
		},
		TokenOut:  types.Token{},
		AmountIn:  amount,
		AmountOut: minReturn,
		Params: map[string]interface{}{
			"function": "unoswapTo",
		},
	}

	return []types.DecodedAction{action}, nil
}

func (d *OneInchDecoder) decodeUniswapV3Swap(calldata []byte) ([]types.DecodedAction, error) {
	aggregatorABI, err := GetAggregatorABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	method, exists := aggregatorABI.Methods["uniswapV3Swap"]
	if !exists {
		return nil, errors.New("uniswapV3Swap method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack uniswapV3Swap parameters: %w", err)
	}

	if len(values) < 3 {
		return nil, errors.New("insufficient parameters for uniswapV3Swap")
	}

	amount, _ := values[0].(*big.Int)
	minReturn, _ := values[1].(*big.Int)

	action := types.DecodedAction{
		Type:      types.SwapAction,
		Protocol:  types.OneInchProtocol,
		TokenIn:   types.Token{},
		TokenOut:  types.Token{},
		AmountIn:  amount,
		AmountOut: minReturn,
		Params: map[string]interface{}{
			"function": "uniswapV3Swap",
		},
	}

	return []types.DecodedAction{action}, nil
}

func (d *OneInchDecoder) decodeUniswapV3SwapTo(calldata []byte) ([]types.DecodedAction, error) {
	aggregatorABI, err := GetAggregatorABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	method, exists := aggregatorABI.Methods["uniswapV3SwapTo"]
	if !exists {
		return nil, errors.New("uniswapV3SwapTo method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack uniswapV3SwapTo parameters: %w", err)
	}

	if len(values) < 4 {
		return nil, errors.New("insufficient parameters for uniswapV3SwapTo")
	}

	amount, _ := values[1].(*big.Int)
	minReturn, _ := values[2].(*big.Int)

	action := types.DecodedAction{
		Type:      types.SwapAction,
		Protocol:  types.OneInchProtocol,
		TokenIn:   types.Token{},
		TokenOut:  types.Token{},
		AmountIn:  amount,
		AmountOut: minReturn,
		Params: map[string]interface{}{
			"function": "uniswapV3SwapTo",
		},
	}

	return []types.DecodedAction{action}, nil
}

func extractSwapDescription(v interface{}) (*SwapDescription, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, errors.New("expected struct for SwapDescription")
	}

	desc := &SwapDescription{}

	// Try to extract fields by name or index
	fieldNames := []string{"SrcToken", "DstToken", "SrcReceiver", "DstReceiver", "Amount", "MinReturnAmount", "Flags"}
	for i := 0; i < rv.NumField() && i < len(fieldNames); i++ {
		field := rv.Field(i)
		if !field.CanInterface() {
			continue
		}

		switch i {
		case 0:
			if addr, ok := field.Interface().(common.Address); ok {
				desc.SrcToken = addr
			}
		case 1:
			if addr, ok := field.Interface().(common.Address); ok {
				desc.DstToken = addr
			}
		case 2:
			if addr, ok := field.Interface().(common.Address); ok {
				desc.SrcReceiver = addr
			}
		case 3:
			if addr, ok := field.Interface().(common.Address); ok {
				desc.DstReceiver = addr
			}
		case 4:
			if val, ok := field.Interface().(*big.Int); ok {
				desc.Amount = val
			}
		case 5:
			if val, ok := field.Interface().(*big.Int); ok {
				desc.MinReturnAmount = val
			}
		case 6:
			if val, ok := field.Interface().(*big.Int); ok {
				desc.Flags = val
			}
		}
	}

	// Set defaults for nil values
	if desc.Amount == nil {
		desc.Amount = big.NewInt(0)
	}
	if desc.MinReturnAmount == nil {
		desc.MinReturnAmount = big.NewInt(0)
	}
	if desc.Flags == nil {
		desc.Flags = big.NewInt(0)
	}

	return desc, nil
}

// IsRouterAddress checks if the address is a known 1inch router
func IsRouterAddress(addr common.Address) bool {
	addrLower := strings.ToLower(addr.Hex())
	for knownAddr := range KnownRouterAddresses {
		if strings.ToLower(knownAddr.Hex()) == addrLower {
			return true
		}
	}
	return false
}
