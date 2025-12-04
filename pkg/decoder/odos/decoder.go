// Package odos implements decoder for Odos router transactions
package odos

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

// OdosDecoder implements the decoder interface for Odos
type OdosDecoder struct {
	decoder.BaseDecoder
}

// NewOdosDecoder creates a new Odos decoder
func NewOdosDecoder() *OdosDecoder {
	return &OdosDecoder{
		BaseDecoder: decoder.BaseDecoder{
			ProtocolType: types.OdosProtocol,
			Name:         "Odos Router",
		},
	}
}

// Matches checks if a transaction should be decoded by this decoder
func (d *OdosDecoder) Matches(tx *ethtypes.Transaction, toAddress string) bool {
	to := common.HexToAddress(toAddress)
	if _, exists := KnownRouterAddresses[to]; !exists {
		return false
	}
	return decoder.MatchesSignature(tx, FunctionSignatures)
}

// Protocol returns the protocol type
func (d *OdosDecoder) Protocol() types.ProtocolType {
	return types.OdosProtocol
}

// Decode decodes the transaction and returns the decoded actions
func (d *OdosDecoder) Decode(tx *ethtypes.Transaction, toAddress string) ([]types.DecodedAction, error) {
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
	case "swapCompact":
		return d.decodeSwapCompact(calldata)
	case "swapMulti":
		return d.decodeSwapMulti(calldata)
	case "swapMultiCompact":
		return d.decodeSwapMultiCompact(calldata)
	default:
		return nil, fmt.Errorf("unimplemented function: %s", funcName)
	}
}

// SwapTokenInfo represents the token info for Odos swap
type SwapTokenInfo struct {
	InputToken     common.Address
	InputAmount    *big.Int
	InputReceiver  common.Address
	OutputToken    common.Address
	OutputQuote    *big.Int
	OutputMin      *big.Int
	OutputReceiver common.Address
}

func (d *OdosDecoder) decodeSwap(calldata []byte) ([]types.DecodedAction, error) {
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

	// Extract SwapTokenInfo from values[0]
	tokenInfo, err := extractSwapTokenInfo(values[0])
	if err != nil {
		return nil, fmt.Errorf("failed to extract token info: %w", err)
	}

	action := types.DecodedAction{
		Type:     types.SwapAction,
		Protocol: types.OdosProtocol,
		TokenIn: types.Token{
			Address: tokenInfo.InputToken,
		},
		TokenOut: types.Token{
			Address: tokenInfo.OutputToken,
		},
		AmountIn:  tokenInfo.InputAmount,
		AmountOut: tokenInfo.OutputMin,
		Params: map[string]interface{}{
			"function":       "swap",
			"outputQuote":    tokenInfo.OutputQuote.String(),
			"inputReceiver":  tokenInfo.InputReceiver.Hex(),
			"outputReceiver": tokenInfo.OutputReceiver.Hex(),
		},
	}

	return []types.DecodedAction{action}, nil
}

func (d *OdosDecoder) decodeSwapMulti(calldata []byte) ([]types.DecodedAction, error) {
	routerABI, err := GetRouterABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	method, exists := routerABI.Methods["swapMulti"]
	if !exists {
		return nil, errors.New("swapMulti method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack swapMulti parameters: %w", err)
	}

	if len(values) < 3 {
		return nil, errors.New("insufficient parameters for swapMulti")
	}

	// For swapMulti, we return a generic action since it involves multiple tokens
	action := types.DecodedAction{
		Type:      types.SwapAction,
		Protocol:  types.OdosProtocol,
		TokenIn:   types.Token{},
		TokenOut:  types.Token{},
		AmountIn:  big.NewInt(0),
		AmountOut: big.NewInt(0),
		Params: map[string]interface{}{
			"function": "swapMulti",
			"isMulti":  true,
		},
	}

	// Try to extract first input and output if available
	if inputs, ok := values[0].([]struct {
		TokenAddress common.Address
		AmountIn     *big.Int
		Receiver     common.Address
	}); ok && len(inputs) > 0 {
		action.TokenIn.Address = inputs[0].TokenAddress
		action.AmountIn = inputs[0].AmountIn
	}

	if outputs, ok := values[1].([]struct {
		TokenAddress  common.Address
		RelativeValue *big.Int
		Receiver      common.Address
	}); ok && len(outputs) > 0 {
		action.TokenOut.Address = outputs[0].TokenAddress
	}

	return []types.DecodedAction{action}, nil
}

// decodeSwapCompact decodes the compact swap format used on L2s
// The compact format encodes addresses using 2-byte references to a cached address list
// Format: selector(4) + inputToken + outputToken + inputAmount(variable) + outputQuote + slippage + executor + inputReceiver + outputReceiver + referralCode + pathDefinition
func (d *OdosDecoder) decodeSwapCompact(calldata []byte) ([]types.DecodedAction, error) {
	if len(calldata) < 8 {
		return nil, errors.New("calldata too short for swapCompact")
	}

	// Skip the 4-byte selector
	data := calldata[4:]
	pos := 0

	// Read input token address
	inputToken, bytesRead, err := readCompactAddress(data, pos)
	if err != nil {
		return nil, fmt.Errorf("failed to read input token: %w", err)
	}
	pos += bytesRead

	// Read output token address
	outputToken, bytesRead, err := readCompactAddress(data, pos)
	if err != nil {
		return nil, fmt.Errorf("failed to read output token: %w", err)
	}
	pos += bytesRead

	// Read input amount (variable length encoding)
	inputAmount, bytesRead, err := readCompactAmount(data, pos)
	if err != nil {
		return nil, fmt.Errorf("failed to read input amount: %w", err)
	}
	pos += bytesRead

	// Read output quote (variable length)
	outputQuote, bytesRead, err := readCompactAmount(data, pos)
	if err != nil {
		return nil, fmt.Errorf("failed to read output quote: %w", err)
	}
	pos += bytesRead

	// Read slippage tolerance (variable length) - we skip for now
	_, bytesRead, err = readCompactAmount(data, pos)
	if err != nil {
		return nil, fmt.Errorf("failed to read slippage: %w", err)
	}
	pos += bytesRead

	// We have enough info to create a basic action
	action := types.DecodedAction{
		Type:     types.SwapAction,
		Protocol: types.OdosProtocol,
		TokenIn: types.Token{
			Address: inputToken,
		},
		TokenOut: types.Token{
			Address: outputToken,
		},
		AmountIn:  inputAmount,
		AmountOut: outputQuote, // Using quote as expected output
		Params: map[string]interface{}{
			"function": "swapCompact",
		},
	}

	return []types.DecodedAction{action}, nil
}

// decodeSwapMultiCompact decodes compact multi-swap format
func (d *OdosDecoder) decodeSwapMultiCompact(calldata []byte) ([]types.DecodedAction, error) {
	if len(calldata) < 8 {
		return nil, errors.New("calldata too short for swapMultiCompact")
	}

	// For now, return a generic multi-swap action
	// Full decoding would require parsing variable-length arrays
	action := types.DecodedAction{
		Type:      types.SwapAction,
		Protocol:  types.OdosProtocol,
		TokenIn:   types.Token{},
		TokenOut:  types.Token{},
		AmountIn:  big.NewInt(0),
		AmountOut: big.NewInt(0),
		Params: map[string]interface{}{
			"function": "swapMultiCompact",
			"isMulti":  true,
		},
	}

	return []types.DecodedAction{action}, nil
}

// WETH address on Arbitrum - used when compact format specifies native token
var WETHAddress = common.HexToAddress("0x82aF49447D8a07e3bd95BD0d56f35241523fBab1")

// Common token addresses on Arbitrum for Odos cache lookup
// These are the most commonly used tokens that Odos caches for gas efficiency
var commonTokenCache = map[uint16]common.Address{
	0x0000: WETHAddress,                                                       // Native ETH -> WETH
	0x0002: common.HexToAddress("0xaf88d065e77c8cC2239327C5EDb3A432268e5831"), // USDC
	0x0003: common.HexToAddress("0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9"), // USDT
	0x0004: common.HexToAddress("0xFF970A61A04b1cA14834A43f5dE4533eBDDB5CC8"), // USDC.e (bridged)
	0x0005: common.HexToAddress("0xDA10009cBd5D07dd0CeCc66161FC93D7c9000da1"), // DAI
	0x0006: common.HexToAddress("0x2f2a2543B76A4166549F7aaB2e75Bef0aefC5B0f"), // WBTC
	0x0007: common.HexToAddress("0x912CE59144191C1204E64559FE8253a0e49E6548"), // ARB
	0x000a: common.HexToAddress("0xfc5A1A6EB076a2C7aD06eD22C90d7E710E35ad0a"), // GMX
}

// readCompactAddress reads an address in Odos compact format
// Returns: address, bytes consumed, error
func readCompactAddress(data []byte, pos int) (common.Address, int, error) {
	if pos+2 > len(data) {
		return common.Address{}, 0, errors.New("insufficient data for address marker")
	}

	marker := (uint16(data[pos]) << 8) | uint16(data[pos+1])

	// 0x0001 = full 20-byte address follows
	if marker == 0x0001 {
		if pos+2+20 > len(data) {
			return common.Address{}, 0, errors.New("insufficient data for full address")
		}
		var addr common.Address
		copy(addr[:], data[pos+2:pos+22])
		return addr, 22, nil
	}

	// Check common token cache (includes 0x0000 for native ETH -> WETH)
	if addr, exists := commonTokenCache[marker]; exists {
		return addr, 2, nil
	}

	// Other values = index into cached address list (protocol-specific)
	// Return a placeholder address with the cache index encoded
	// This lets us identify it as a cached address while preserving the index
	var addr common.Address
	addr[0] = 0xCA // "CA" prefix to indicate CAched address
	addr[1] = 0xCE
	addr[18] = data[pos]
	addr[19] = data[pos+1]
	return addr, 2, nil
}

// readCompactAmount reads a variable-length encoded amount
// First byte indicates the number of bytes to read for the amount
func readCompactAmount(data []byte, pos int) (*big.Int, int, error) {
	if pos >= len(data) {
		return nil, 0, errors.New("insufficient data for amount length")
	}

	// First byte indicates number of bytes for the amount
	numBytes := int(data[pos])
	if numBytes == 0 {
		return big.NewInt(0), 1, nil
	}
	if numBytes > 32 {
		return nil, 0, fmt.Errorf("invalid amount length: %d (max 32 bytes)", numBytes)
	}

	if pos+1+numBytes > len(data) {
		return nil, 0, errors.New("insufficient data for amount value")
	}

	amount := new(big.Int).SetBytes(data[pos+1 : pos+1+numBytes])
	return amount, 1 + numBytes, nil
}

func (d *OdosDecoder) decodeSwapRouterTokensForTokens(calldata []byte) ([]types.DecodedAction, error) {
	routerABI, err := GetRouterABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	method, exists := routerABI.Methods["swapRouterTokensForTokens"]
	if !exists {
		return nil, errors.New("swapRouterTokensForTokens method not found in ABI")
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack swapRouterTokensForTokens parameters: %w", err)
	}

	if len(values) < 5 {
		return nil, errors.New("insufficient parameters for swapRouterTokensForTokens")
	}

	tokenIn, _ := values[0].(common.Address)
	amountIn, _ := values[1].(*big.Int)
	tokenOut, _ := values[2].(common.Address)
	amountOutMin, _ := values[3].(*big.Int)
	to, _ := values[4].(common.Address)

	action := types.DecodedAction{
		Type:     types.SwapAction,
		Protocol: types.OdosProtocol,
		TokenIn: types.Token{
			Address: tokenIn,
		},
		TokenOut: types.Token{
			Address: tokenOut,
		},
		AmountIn:  amountIn,
		AmountOut: amountOutMin,
		Params: map[string]interface{}{
			"function":  "swapRouterTokensForTokens",
			"recipient": to.Hex(),
		},
	}

	return []types.DecodedAction{action}, nil
}

func extractSwapTokenInfo(v interface{}) (*SwapTokenInfo, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, errors.New("expected struct for SwapTokenInfo")
	}

	info := &SwapTokenInfo{
		InputAmount: big.NewInt(0),
		OutputQuote: big.NewInt(0),
		OutputMin:   big.NewInt(0),
	}

	for i := 0; i < rv.NumField() && i < 7; i++ {
		field := rv.Field(i)
		if !field.CanInterface() {
			continue
		}

		switch i {
		case 0:
			if addr, ok := field.Interface().(common.Address); ok {
				info.InputToken = addr
			}
		case 1:
			if val, ok := field.Interface().(*big.Int); ok {
				info.InputAmount = val
			}
		case 2:
			if addr, ok := field.Interface().(common.Address); ok {
				info.InputReceiver = addr
			}
		case 3:
			if addr, ok := field.Interface().(common.Address); ok {
				info.OutputToken = addr
			}
		case 4:
			if val, ok := field.Interface().(*big.Int); ok {
				info.OutputQuote = val
			}
		case 5:
			if val, ok := field.Interface().(*big.Int); ok {
				info.OutputMin = val
			}
		case 6:
			if addr, ok := field.Interface().(common.Address); ok {
				info.OutputReceiver = addr
			}
		}
	}

	return info, nil
}

// IsRouterAddress checks if the address is a known Odos router
func IsRouterAddress(addr common.Address) bool {
	addrLower := strings.ToLower(addr.Hex())
	for knownAddr := range KnownRouterAddresses {
		if strings.ToLower(knownAddr.Hex()) == addrLower {
			return true
		}
	}
	return false
}
