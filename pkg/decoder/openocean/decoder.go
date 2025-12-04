// Package openocean implements decoder for OpenOcean exchange transactions
package openocean

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

// OpenOceanDecoder implements the decoder interface for OpenOcean
type OpenOceanDecoder struct {
	decoder.BaseDecoder
}

// NewOpenOceanDecoder creates a new OpenOcean decoder
func NewOpenOceanDecoder() *OpenOceanDecoder {
	return &OpenOceanDecoder{
		BaseDecoder: decoder.BaseDecoder{
			ProtocolType: types.OpenOceanProtocol,
			Name:         "OpenOcean Exchange",
		},
	}
}

// Matches checks if a transaction should be decoded by this decoder
func (d *OpenOceanDecoder) Matches(tx *ethtypes.Transaction, toAddress string) bool {
	to := common.HexToAddress(toAddress)
	if _, exists := KnownRouterAddresses[to]; !exists {
		return false
	}
	return decoder.MatchesSignature(tx, FunctionSignatures)
}

// Protocol returns the protocol type
func (d *OpenOceanDecoder) Protocol() types.ProtocolType {
	return types.OpenOceanProtocol
}

// Decode decodes the transaction and returns the decoded actions
func (d *OpenOceanDecoder) Decode(tx *ethtypes.Transaction, toAddress string) ([]types.DecodedAction, error) {
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
	case "callUniswapTo":
		return d.decodeCallUniswapTo(calldata)
	default:
		return nil, fmt.Errorf("unimplemented function: %s", funcName)
	}
}

// SwapDescription represents the swap parameters for OpenOcean swap
type SwapDescription struct {
	SrcToken         common.Address
	DstToken         common.Address
	SrcReceiver      common.Address
	DstReceiver      common.Address
	Amount           *big.Int
	MinReturnAmount  *big.Int
	GuaranteedAmount *big.Int
	Flags            *big.Int
	Referrer         common.Address
}

func (d *OpenOceanDecoder) decodeSwap(calldata []byte) ([]types.DecodedAction, error) {
	exchangeABI, err := GetExchangeABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	method, exists := exchangeABI.Methods["swap"]
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
		Protocol: types.OpenOceanProtocol,
		TokenIn: types.Token{
			Address: desc.SrcToken,
		},
		TokenOut: types.Token{
			Address: desc.DstToken,
		},
		AmountIn:  desc.Amount,
		AmountOut: desc.MinReturnAmount,
		Params: map[string]interface{}{
			"function":         "swap",
			"srcReceiver":      desc.SrcReceiver.Hex(),
			"dstReceiver":      desc.DstReceiver.Hex(),
			"guaranteedAmount": desc.GuaranteedAmount.String(),
			"referrer":         desc.Referrer.Hex(),
		},
	}

	return []types.DecodedAction{action}, nil
}

func (d *OpenOceanDecoder) decodeCallUniswapTo(calldata []byte) ([]types.DecodedAction, error) {
	exchangeABI, err := GetExchangeABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	method, exists := exchangeABI.Methods["callUniswapTo"]
	if !exists {
		return nil, errors.New("callUniswapTo method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack callUniswapTo parameters: %w", err)
	}

	if len(values) < 5 {
		return nil, errors.New("insufficient parameters for callUniswapTo")
	}

	srcToken, _ := values[1].(common.Address)
	dstToken, _ := values[2].(common.Address)
	amount, _ := values[3].(*big.Int)
	minReturn, _ := values[4].(*big.Int)

	action := types.DecodedAction{
		Type:     types.SwapAction,
		Protocol: types.OpenOceanProtocol,
		TokenIn: types.Token{
			Address: srcToken,
		},
		TokenOut: types.Token{
			Address: dstToken,
		},
		AmountIn:  amount,
		AmountOut: minReturn,
		Params: map[string]interface{}{
			"function": "callUniswapTo",
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

	desc := &SwapDescription{
		Amount:           big.NewInt(0),
		MinReturnAmount:  big.NewInt(0),
		GuaranteedAmount: big.NewInt(0),
		Flags:            big.NewInt(0),
	}

	for i := 0; i < rv.NumField() && i < 10; i++ {
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
				desc.GuaranteedAmount = val
			}
		case 7:
			if val, ok := field.Interface().(*big.Int); ok {
				desc.Flags = val
			}
		case 8:
			if addr, ok := field.Interface().(common.Address); ok {
				desc.Referrer = addr
			}
		}
	}

	return desc, nil
}

// IsRouterAddress checks if the address is a known OpenOcean router
func IsRouterAddress(addr common.Address) bool {
	addrLower := strings.ToLower(addr.Hex())
	for knownAddr := range KnownRouterAddresses {
		if strings.ToLower(knownAddr.Hex()) == addrLower {
			return true
		}
	}
	return false
}
