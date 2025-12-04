// Package kyberswap implements decoder for KyberSwap router transactions
package kyberswap

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

// KyberSwapDecoder implements the decoder interface for KyberSwap
type KyberSwapDecoder struct {
	decoder.BaseDecoder
}

// NewKyberSwapDecoder creates a new KyberSwap decoder
func NewKyberSwapDecoder() *KyberSwapDecoder {
	return &KyberSwapDecoder{
		BaseDecoder: decoder.BaseDecoder{
			ProtocolType: types.KyberElasticProtocol,
			Name:         "KyberSwap",
		},
	}
}

// Matches checks if a transaction should be decoded by this decoder
func (d *KyberSwapDecoder) Matches(tx *ethtypes.Transaction, toAddress string) bool {
	to := common.HexToAddress(toAddress)
	if _, exists := KnownRouterAddresses[to]; !exists {
		return false
	}
	return decoder.MatchesSignature(tx, FunctionSignatures)
}

// Protocol returns the protocol type
func (d *KyberSwapDecoder) Protocol() types.ProtocolType {
	return types.KyberElasticProtocol
}

// Decode decodes the transaction and returns the decoded actions
func (d *KyberSwapDecoder) Decode(tx *ethtypes.Transaction, toAddress string) ([]types.DecodedAction, error) {
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
	default:
		return nil, fmt.Errorf("unimplemented function: %s", funcName)
	}
}

// SwapDescription represents the swap parameters for KyberSwap MetaAggregationRouterV2
// This is the inner desc field within SwapExecutionParams
type SwapDescription struct {
	SrcToken        common.Address
	DstToken        common.Address
	DstReceiver     common.Address
	Amount          *big.Int
	MinReturnAmount *big.Int
	Flags           *big.Int
}

func (d *KyberSwapDecoder) decodeSwap(calldata []byte) ([]types.DecodedAction, error) {
	routerABI, err := GetRouterABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	method, exists := routerABI.Methods["swap"]
	if !exists {
		return nil, errors.New("swap method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack swap parameters: %w", err)
	}

	if len(values) < 1 {
		return nil, errors.New("insufficient parameters for swap")
	}

	// values[0] is the SwapExecutionParams tuple
	// It contains: callTarget, approveTarget, targetData, desc (SwapDescription), clientData
	desc, err := extractSwapDescription(values[0])
	if err != nil {
		return nil, fmt.Errorf("failed to extract swap description: %w", err)
	}

	action := types.DecodedAction{
		Type:     types.SwapAction,
		Protocol: types.KyberElasticProtocol,
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
			"dstReceiver": desc.DstReceiver.Hex(),
		},
	}

	return []types.DecodedAction{action}, nil
}

// extractSwapDescription extracts swap details from SwapExecutionParams
// SwapExecutionParams contains: callTarget, approveTarget, targetData, desc (SwapDescription), clientData
// SwapDescription contains: srcToken, dstToken, srcReceivers[], srcAmounts[], feeReceivers[], feeAmounts[], dstReceiver, amount, minReturnAmount, flags, permit
func extractSwapDescription(v interface{}) (*SwapDescription, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, errors.New("expected struct for SwapExecutionParams")
	}

	// SwapExecutionParams has 5 fields, we need field index 3 (desc)
	if rv.NumField() < 4 {
		return nil, errors.New("SwapExecutionParams has insufficient fields")
	}

	// Get the desc field (index 3)
	descField := rv.Field(3)
	if descField.Kind() == reflect.Ptr {
		descField = descField.Elem()
	}

	if descField.Kind() != reflect.Struct {
		return nil, errors.New("expected struct for SwapDescription")
	}

	desc := &SwapDescription{
		Amount:          big.NewInt(0),
		MinReturnAmount: big.NewInt(0),
		Flags:           big.NewInt(0),
	}

	// SwapDescription field mapping:
	// 0: srcToken, 1: dstToken, 2: srcReceivers[], 3: srcAmounts[],
	// 4: feeReceivers[], 5: feeAmounts[], 6: dstReceiver, 7: amount,
	// 8: minReturnAmount, 9: flags, 10: permit

	for i := 0; i < descField.NumField(); i++ {
		field := descField.Field(i)
		if !field.CanInterface() {
			continue
		}

		switch i {
		case 0: // srcToken
			if addr, ok := field.Interface().(common.Address); ok {
				desc.SrcToken = addr
			}
		case 1: // dstToken
			if addr, ok := field.Interface().(common.Address); ok {
				desc.DstToken = addr
			}
		case 6: // dstReceiver
			if addr, ok := field.Interface().(common.Address); ok {
				desc.DstReceiver = addr
			}
		case 7: // amount
			if val, ok := field.Interface().(*big.Int); ok {
				desc.Amount = val
			}
		case 8: // minReturnAmount
			if val, ok := field.Interface().(*big.Int); ok {
				desc.MinReturnAmount = val
			}
		case 9: // flags
			if val, ok := field.Interface().(*big.Int); ok {
				desc.Flags = val
			}
		}
	}

	return desc, nil
}

// IsRouterAddress checks if the address is a known KyberSwap router
func IsRouterAddress(addr common.Address) bool {
	addrLower := strings.ToLower(addr.Hex())
	for knownAddr := range KnownRouterAddresses {
		if strings.ToLower(knownAddr.Hex()) == addrLower {
			return true
		}
	}
	return false
}
