// Package zerox implements decoder for 0x Exchange Proxy transactions
package zerox

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoder"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
)

// ZeroXDecoder implements the decoder interface for 0x Exchange Proxy
type ZeroXDecoder struct {
	decoder.BaseDecoder
}

// NewZeroXDecoder creates a new 0x decoder
func NewZeroXDecoder() *ZeroXDecoder {
	return &ZeroXDecoder{
		BaseDecoder: decoder.BaseDecoder{
			ProtocolType: types.ZeroXProtocol,
			Name:         "0x",
		},
	}
}

// Matches checks if a transaction should be decoded by this decoder
func (d *ZeroXDecoder) Matches(tx *ethtypes.Transaction, toAddress string) bool {
	addr := common.HexToAddress(toAddress)

	if _, exists := KnownRouterAddresses[addr]; !exists {
		return false
	}

	return decoder.MatchesSignature(tx, FunctionSignatures)
}

// Decode decodes the transaction data and returns the decoded actions
func (d *ZeroXDecoder) Decode(tx *ethtypes.Transaction, toAddress string) ([]types.DecodedAction, error) {
	calldata := tx.Data()
	if len(calldata) < 4 {
		return nil, errors.New("calldata too short")
	}

	signature := "0x" + common.Bytes2Hex(calldata[:4])

	funcName, exists := FunctionSignatureMap[signature]
	if !exists {
		return nil, fmt.Errorf("unknown 0x function signature: %s", signature)
	}

	switch funcName {
	case "sellToUniswap":
		return d.decodeSellToUniswap(calldata)
	case "transformERC20":
		return d.decodeTransformERC20(calldata)
	case "sellToLiquidityProvider":
		return d.decodeSellToLiquidityProvider(calldata)
	default:
		return nil, fmt.Errorf("unhandled function: %s", funcName)
	}
}

// Protocol returns the protocol this decoder handles
func (d *ZeroXDecoder) Protocol() types.ProtocolType {
	return d.ProtocolType
}

// decodeSellToUniswap decodes sellToUniswap function
func (d *ZeroXDecoder) decodeSellToUniswap(calldata []byte) ([]types.DecodedAction, error) {
	proxyABI, err := GetExchangeProxyABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get 0x ABI: %w", err)
	}

	method, exists := proxyABI.Methods["sellToUniswap"]
	if !exists {
		return nil, errors.New("sellToUniswap method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack sellToUniswap parameters: %w", err)
	}

	if len(values) < 4 {
		return nil, errors.New("insufficient parameters for sellToUniswap")
	}

	tokens, ok := values[0].([]common.Address)
	if !ok {
		return nil, errors.New("failed to extract tokens")
	}

	sellAmount, ok := values[1].(*big.Int)
	if !ok {
		return nil, errors.New("failed to extract sellAmount")
	}

	minBuyAmount, ok := values[2].(*big.Int)
	if !ok {
		return nil, errors.New("failed to extract minBuyAmount")
	}

	isSushi, ok := values[3].(bool)
	if !ok {
		isSushi = false
	}

	if len(tokens) < 2 {
		return nil, errors.New("path must have at least 2 tokens")
	}

	// Create actions for each hop
	var actions []types.DecodedAction
	totalHops := len(tokens) - 1
	for i := 0; i < totalHops; i++ {
		var hopAmountIn, hopAmountOut *big.Int
		if i == 0 {
			hopAmountIn = sellAmount
		}
		if i == totalHops-1 {
			hopAmountOut = minBuyAmount
		}

		protocol := types.UniswapV2Protocol
		if isSushi {
			protocol = types.SushiSwapProtocol
		}

		action := types.DecodedAction{
			Type:      types.SwapAction,
			Protocol:  protocol,
			TokenIn:   types.Token{Address: tokens[i]},
			TokenOut:  types.Token{Address: tokens[i+1]},
			AmountIn:  hopAmountIn,
			AmountOut: hopAmountOut,
			Params: map[string]interface{}{
				"function":     "sellToUniswap",
				"isSushi":      isSushi,
				"minBuyAmount": minBuyAmount,
				"pathIndex":    i,
				"totalHops":    totalHops,
			},
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// decodeTransformERC20 decodes transformERC20 function
func (d *ZeroXDecoder) decodeTransformERC20(calldata []byte) ([]types.DecodedAction, error) {
	proxyABI, err := GetExchangeProxyABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get 0x ABI: %w", err)
	}

	method, exists := proxyABI.Methods["transformERC20"]
	if !exists {
		return nil, errors.New("transformERC20 method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack transformERC20 parameters: %w", err)
	}

	if len(values) < 4 {
		return nil, errors.New("insufficient parameters for transformERC20")
	}

	inputToken, ok := values[0].(common.Address)
	if !ok {
		return nil, errors.New("failed to extract inputToken")
	}

	outputToken, ok := values[1].(common.Address)
	if !ok {
		return nil, errors.New("failed to extract outputToken")
	}

	inputTokenAmount, ok := values[2].(*big.Int)
	if !ok {
		return nil, errors.New("failed to extract inputTokenAmount")
	}

	minOutputTokenAmount, ok := values[3].(*big.Int)
	if !ok {
		return nil, errors.New("failed to extract minOutputTokenAmount")
	}

	action := types.DecodedAction{
		Type:      types.SwapAction,
		Protocol:  types.ZeroXProtocol,
		TokenIn:   types.Token{Address: inputToken},
		TokenOut:  types.Token{Address: outputToken},
		AmountIn:  inputTokenAmount,
		AmountOut: minOutputTokenAmount,
		Params: map[string]interface{}{
			"function":             "transformERC20",
			"minOutputTokenAmount": minOutputTokenAmount,
		},
	}

	return []types.DecodedAction{action}, nil
}

// decodeSellToLiquidityProvider decodes sellToLiquidityProvider function
func (d *ZeroXDecoder) decodeSellToLiquidityProvider(calldata []byte) ([]types.DecodedAction, error) {
	proxyABI, err := GetExchangeProxyABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get 0x ABI: %w", err)
	}

	method, exists := proxyABI.Methods["sellToLiquidityProvider"]
	if !exists {
		return nil, errors.New("sellToLiquidityProvider method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack sellToLiquidityProvider parameters: %w", err)
	}

	if len(values) < 6 {
		return nil, errors.New("insufficient parameters for sellToLiquidityProvider")
	}

	inputToken := extractAddress(values[0])
	outputToken := extractAddress(values[1])
	provider := extractAddress(values[2])
	recipient := extractAddress(values[3])
	sellAmount := extractBigInt(values[4])
	minBuyAmount := extractBigInt(values[5])

	action := types.DecodedAction{
		Type:      types.SwapAction,
		Protocol:  types.ZeroXProtocol,
		TokenIn:   types.Token{Address: inputToken},
		TokenOut:  types.Token{Address: outputToken},
		AmountIn:  sellAmount,
		AmountOut: minBuyAmount,
		Params: map[string]interface{}{
			"function":     "sellToLiquidityProvider",
			"provider":     provider,
			"recipient":    recipient,
			"minBuyAmount": minBuyAmount,
		},
	}

	return []types.DecodedAction{action}, nil
}

func extractAddress(v interface{}) common.Address {
	if addr, ok := v.(common.Address); ok {
		return addr
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Array && rv.Len() == 20 {
		var addr common.Address
		for i := 0; i < 20; i++ {
			addr[i] = byte(rv.Index(i).Uint())
		}
		return addr
	}
	return common.Address{}
}

func extractBigInt(v interface{}) *big.Int {
	if bi, ok := v.(*big.Int); ok {
		return bi
	}
	return big.NewInt(0)
}
