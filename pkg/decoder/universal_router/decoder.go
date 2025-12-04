// Package universal_router implements decoder for Uniswap Universal Router transactions
package universal_router

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoder"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
)

// UniversalRouterDecoder implements the decoder interface for Uniswap Universal Router
type UniversalRouterDecoder struct {
	decoder.BaseDecoder
}

// NewUniversalRouterDecoder creates a new Universal Router decoder
func NewUniversalRouterDecoder() *UniversalRouterDecoder {
	return &UniversalRouterDecoder{
		BaseDecoder: decoder.BaseDecoder{
			ProtocolType: types.UniswapV3Protocol, // Uses V3 protocol type since it routes to V3 pools
			Name:         "UniversalRouter",
		},
	}
}

// Matches checks if a transaction should be decoded by this decoder
func (d *UniversalRouterDecoder) Matches(tx *ethtypes.Transaction, toAddress string) bool {
	addr := common.HexToAddress(toAddress)

	// Check if the destination address is a known Universal Router
	if _, exists := KnownRouterAddresses[addr]; !exists {
		return false
	}

	// Check if the transaction matches any Universal Router function signatures
	return decoder.MatchesSignature(tx, FunctionSignatures)
}

// Decode decodes the transaction data and returns the decoded actions
func (d *UniversalRouterDecoder) Decode(tx *ethtypes.Transaction, toAddress string) ([]types.DecodedAction, error) {
	calldata := tx.Data()
	if len(calldata) < 4 {
		return nil, errors.New("calldata too short")
	}

	// Extract function signature
	signature := "0x" + common.Bytes2Hex(calldata[:4])

	funcName, exists := FunctionSignatureMap[signature]
	if !exists {
		return nil, fmt.Errorf("unknown Universal Router function signature: %s", signature)
	}

	switch funcName {
	case "execute", "executeNoDeadline":
		return d.decodeExecute(calldata, funcName == "executeNoDeadline")
	default:
		return nil, fmt.Errorf("unhandled function: %s", funcName)
	}
}

// Protocol returns the protocol this decoder handles
func (d *UniversalRouterDecoder) Protocol() types.ProtocolType {
	return d.ProtocolType
}

// decodeExecute decodes the execute function
func (d *UniversalRouterDecoder) decodeExecute(calldata []byte, noDeadline bool) ([]types.DecodedAction, error) {
	routerABI, err := GetRouterABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get router ABI: %w", err)
	}

	var methodName string
	if noDeadline {
		methodName = "execute0" // go-ethereum names overloaded functions with index
	} else {
		methodName = "execute"
	}

	method, exists := routerABI.Methods[methodName]
	if !exists {
		// Try the other variant
		method, exists = routerABI.Methods["execute"]
		if !exists {
			return nil, errors.New("execute method not found in ABI")
		}
	}

	values, err := method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack execute parameters: %w", err)
	}

	if len(values) < 2 {
		return nil, errors.New("insufficient parameters for execute")
	}

	commands, ok := values[0].([]byte)
	if !ok {
		return nil, errors.New("failed to extract commands from execute")
	}

	inputs, ok := values[1].([][]byte)
	if !ok {
		return nil, errors.New("failed to extract inputs from execute")
	}

	if len(commands) != len(inputs) {
		return nil, fmt.Errorf("commands length (%d) != inputs length (%d)", len(commands), len(inputs))
	}

	var allActions []types.DecodedAction

	for i, cmd := range commands {
		// Mask off the flag bit (0x80) which indicates whether to allow revert
		cmdType := cmd & 0x3f

		actions, err := d.decodeCommand(cmdType, inputs[i])
		if err != nil {
			// Skip commands we can't decode but continue with others
			continue
		}

		for j := range actions {
			if actions[j].Params == nil {
				actions[j].Params = make(map[string]interface{})
			}
			actions[j].Params["commandIndex"] = i
			actions[j].Params["commandType"] = cmdType
		}

		allActions = append(allActions, actions...)
	}

	return allActions, nil
}

// decodeCommand decodes a single Universal Router command
func (d *UniversalRouterDecoder) decodeCommand(cmdType byte, input []byte) ([]types.DecodedAction, error) {
	switch cmdType {
	case V3_SWAP_EXACT_IN:
		return d.decodeV3SwapExactIn(input)
	case V3_SWAP_EXACT_OUT:
		return d.decodeV3SwapExactOut(input)
	case V2_SWAP_EXACT_IN:
		return d.decodeV2SwapExactIn(input)
	case V2_SWAP_EXACT_OUT:
		return d.decodeV2SwapExactOut(input)
	case WRAP_ETH, UNWRAP_WETH, SWEEP, TRANSFER, PAY_PORTION,
		PERMIT2_PERMIT, PERMIT2_TRANSFER_FROM, PERMIT2_PERMIT_BATCH, PERMIT2_TRANSFER_BATCH:
		// Skip non-swap commands silently
		return nil, nil
	default:
		// Unknown command, skip
		return nil, nil
	}
}

// decodeV3SwapExactIn decodes V3_SWAP_EXACT_IN command
// Input: (address recipient, uint256 amountIn, uint256 amountOutMin, bytes path, bool payerIsUser)
func (d *UniversalRouterDecoder) decodeV3SwapExactIn(input []byte) ([]types.DecodedAction, error) {
	if len(input) < 160 { // Minimum: 32 + 32 + 32 + 32 (offset) + 32 (path length) = 160
		return nil, errors.New("input too short for V3_SWAP_EXACT_IN")
	}

	// Decode parameters (ABI encoded)
	recipient := common.BytesToAddress(input[12:32])
	amountIn := new(big.Int).SetBytes(input[32:64])
	amountOutMin := new(big.Int).SetBytes(input[64:96])

	// Path offset is at bytes 96-128
	pathOffset := new(big.Int).SetBytes(input[96:128]).Uint64()
	if pathOffset+32 > uint64(len(input)) {
		return nil, errors.New("invalid path offset")
	}

	// Read path length and data
	pathLen := new(big.Int).SetBytes(input[pathOffset : pathOffset+32]).Uint64()
	if pathOffset+32+pathLen > uint64(len(input)) {
		return nil, errors.New("path data out of bounds")
	}
	pathData := input[pathOffset+32 : pathOffset+32+pathLen]

	// Decode V3 path
	tokens, err := decodeV3Path(pathData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode V3 path: %w", err)
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
			hopAmountIn = amountIn
		}
		if i == totalHops-1 {
			hopAmountOut = amountOutMin
		}

		action := types.DecodedAction{
			Type:      types.SwapAction,
			Protocol:  types.UniswapV3Protocol,
			TokenIn:   types.Token{Address: tokens[i]},
			TokenOut:  types.Token{Address: tokens[i+1]},
			AmountIn:  hopAmountIn,
			AmountOut: hopAmountOut,
			Params: map[string]interface{}{
				"function":         "V3_SWAP_EXACT_IN",
				"recipient":        recipient,
				"amountOutMinimum": amountOutMin,
				"pathIndex":        i,
				"totalHops":        totalHops,
			},
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// decodeV3SwapExactOut decodes V3_SWAP_EXACT_OUT command
// Input: (address recipient, uint256 amountOut, uint256 amountInMax, bytes path, bool payerIsUser)
func (d *UniversalRouterDecoder) decodeV3SwapExactOut(input []byte) ([]types.DecodedAction, error) {
	if len(input) < 160 {
		return nil, errors.New("input too short for V3_SWAP_EXACT_OUT")
	}

	recipient := common.BytesToAddress(input[12:32])
	amountOut := new(big.Int).SetBytes(input[32:64])
	amountInMax := new(big.Int).SetBytes(input[64:96])

	pathOffset := new(big.Int).SetBytes(input[96:128]).Uint64()
	if pathOffset+32 > uint64(len(input)) {
		return nil, errors.New("invalid path offset")
	}

	pathLen := new(big.Int).SetBytes(input[pathOffset : pathOffset+32]).Uint64()
	if pathOffset+32+pathLen > uint64(len(input)) {
		return nil, errors.New("path data out of bounds")
	}
	pathData := input[pathOffset+32 : pathOffset+32+pathLen]

	// Decode V3 path (reversed for exact output)
	tokens, err := decodeV3Path(pathData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode V3 path: %w", err)
	}

	if len(tokens) < 2 {
		return nil, errors.New("path must have at least 2 tokens")
	}

	// Create actions - path is reversed for exactOutput
	var actions []types.DecodedAction
	totalHops := len(tokens) - 1
	for i := 0; i < totalHops; i++ {
		var hopAmountIn, hopAmountOut *big.Int
		if i == 0 {
			hopAmountOut = amountOut
		}
		if i == totalHops-1 {
			hopAmountIn = amountInMax
		}

		action := types.DecodedAction{
			Type:      types.SwapAction,
			Protocol:  types.UniswapV3Protocol,
			TokenIn:   types.Token{Address: tokens[i+1]}, // Reversed
			TokenOut:  types.Token{Address: tokens[i]},   // Reversed
			AmountIn:  hopAmountIn,
			AmountOut: hopAmountOut,
			Params: map[string]interface{}{
				"function":        "V3_SWAP_EXACT_OUT",
				"recipient":       recipient,
				"amountInMaximum": amountInMax,
				"pathIndex":       i,
				"totalHops":       totalHops,
			},
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// decodeV2SwapExactIn decodes V2_SWAP_EXACT_IN command
// Input: (address recipient, uint256 amountIn, uint256 amountOutMin, address[] path, bool payerIsUser)
func (d *UniversalRouterDecoder) decodeV2SwapExactIn(input []byte) ([]types.DecodedAction, error) {
	if len(input) < 160 {
		return nil, errors.New("input too short for V2_SWAP_EXACT_IN")
	}

	recipient := common.BytesToAddress(input[12:32])
	amountIn := new(big.Int).SetBytes(input[32:64])
	amountOutMin := new(big.Int).SetBytes(input[64:96])

	// Path is address[] - offset at bytes 96-128
	pathOffset := new(big.Int).SetBytes(input[96:128]).Uint64()
	if pathOffset+32 > uint64(len(input)) {
		return nil, errors.New("invalid path offset")
	}

	pathLen := new(big.Int).SetBytes(input[pathOffset : pathOffset+32]).Uint64()

	var tokens []common.Address
	for i := uint64(0); i < pathLen; i++ {
		start := pathOffset + 32 + (i * 32) + 12
		end := start + 20
		if end > uint64(len(input)) {
			return nil, errors.New("path data out of bounds")
		}
		tokens = append(tokens, common.BytesToAddress(input[start:end]))
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
			hopAmountIn = amountIn
		}
		if i == totalHops-1 {
			hopAmountOut = amountOutMin
		}

		action := types.DecodedAction{
			Type:      types.SwapAction,
			Protocol:  types.UniswapV2Protocol,
			TokenIn:   types.Token{Address: tokens[i]},
			TokenOut:  types.Token{Address: tokens[i+1]},
			AmountIn:  hopAmountIn,
			AmountOut: hopAmountOut,
			Params: map[string]interface{}{
				"function":         "V2_SWAP_EXACT_IN",
				"recipient":        recipient,
				"amountOutMinimum": amountOutMin,
				"pathIndex":        i,
				"totalHops":        totalHops,
			},
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// decodeV2SwapExactOut decodes V2_SWAP_EXACT_OUT command
// Input: (address recipient, uint256 amountOut, uint256 amountInMax, address[] path, bool payerIsUser)
func (d *UniversalRouterDecoder) decodeV2SwapExactOut(input []byte) ([]types.DecodedAction, error) {
	if len(input) < 160 {
		return nil, errors.New("input too short for V2_SWAP_EXACT_OUT")
	}

	recipient := common.BytesToAddress(input[12:32])
	amountOut := new(big.Int).SetBytes(input[32:64])
	amountInMax := new(big.Int).SetBytes(input[64:96])

	pathOffset := new(big.Int).SetBytes(input[96:128]).Uint64()
	if pathOffset+32 > uint64(len(input)) {
		return nil, errors.New("invalid path offset")
	}

	pathLen := new(big.Int).SetBytes(input[pathOffset : pathOffset+32]).Uint64()

	var tokens []common.Address
	for i := uint64(0); i < pathLen; i++ {
		start := pathOffset + 32 + (i * 32) + 12
		end := start + 20
		if end > uint64(len(input)) {
			return nil, errors.New("path data out of bounds")
		}
		tokens = append(tokens, common.BytesToAddress(input[start:end]))
	}

	if len(tokens) < 2 {
		return nil, errors.New("path must have at least 2 tokens")
	}

	// Create actions - V2 exact output path is in reverse order
	var actions []types.DecodedAction
	totalHops := len(tokens) - 1
	for i := totalHops - 1; i >= 0; i-- {
		var hopAmountIn, hopAmountOut *big.Int
		if i == totalHops-1 {
			hopAmountOut = amountOut
		}
		if i == 0 {
			hopAmountIn = amountInMax
		}

		action := types.DecodedAction{
			Type:      types.SwapAction,
			Protocol:  types.UniswapV2Protocol,
			TokenIn:   types.Token{Address: tokens[i]},
			TokenOut:  types.Token{Address: tokens[i+1]},
			AmountIn:  hopAmountIn,
			AmountOut: hopAmountOut,
			Params: map[string]interface{}{
				"function":        "V2_SWAP_EXACT_OUT",
				"recipient":       recipient,
				"amountInMaximum": amountInMax,
				"pathIndex":       totalHops - 1 - i,
				"totalHops":       totalHops,
			},
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// decodeV3Path decodes a Uniswap V3 path (token + fee + token + fee + ... + token)
// Each token is 20 bytes, each fee is 3 bytes
func decodeV3Path(path []byte) ([]common.Address, error) {
	if len(path) < 43 { // Minimum: 20 + 3 + 20 = 43
		return nil, errors.New("path too short")
	}

	var tokens []common.Address

	// First token
	tokens = append(tokens, common.BytesToAddress(path[0:20]))

	// Remaining tokens (after fee)
	pos := 20
	for pos+23 <= len(path) {
		// Skip 3 byte fee
		pos += 3
		// Read 20 byte token
		tokens = append(tokens, common.BytesToAddress(path[pos:pos+20]))
		pos += 20
	}

	return tokens, nil
}
