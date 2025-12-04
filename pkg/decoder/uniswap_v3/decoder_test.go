// Tests for Uniswap V3 decoder
package uniswap_v3

import (
	"encoding/hex"
	"testing"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/testhelpers"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
	"github.com/ethereum/go-ethereum/common"
)

func TestUniswapV3Decoder_Matches(t *testing.T) {
	decoder := NewUniswapV3Decoder()

	// Test with a known Uniswap V3 router address
	uniV3RouterAddr := common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564")

	// Create a transaction with exactInput function signature (0xc04b8d59)
	// This is properly ABI-encoded calldata for exactInput
	calldata, _ := hex.DecodeString("c04b8d59" + "0000000000000000000000000000000000000000000000000000000000000020" + // offset to params tuple
		"00000000000000000000000000000000000000000000000000000000000000a0" + // offset to path within tuple
		"0000000000000000000000001234567890123456789012345678901234567890" + // recipient
		"0000000000000000000000000000000000000000000000000000000065f5e100" + // deadline
		"0000000000000000000000000000000000000000000000000de0b6b3a7640000" + // amountIn (1 ETH)
		"0000000000000000000000000000000000000000000000000000000000000001" + // amountOutMinimum
		"000000000000000000000000000000000000000000000000000000000000002b" + // path length (43 bytes)
		"a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48" + // token0 (USDC)
		"000bb8" + // fee (3000)
		"d533a949740bb3306d11b9d5bc00000000000000" + // token1
		"00000000000000000000000000000000000000") // padding
	tx := testhelpers.CreateTestTransaction(uniV3RouterAddr, calldata, nil)

	// Test that the decoder matches transactions to Uniswap V3 router with proper function signature
	if !decoder.Matches(tx, uniV3RouterAddr.Hex()) {
		t.Error("Expected decoder to match Uniswap V3 transaction with exactInput signature")
	}

	// Test with non-Uniswap V3 address
	nonUniAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	if decoder.Matches(tx, nonUniAddr.Hex()) {
		t.Error("Expected decoder to not match transaction to non-Uniswap V3 address")
	}

	// Test with non-swap function signature
	nonSwapCalldata, _ := hex.DecodeString("12345678" + "0000000000000000000000000000000000000000000000000000000000000000")
	nonSwapTx := testhelpers.CreateTestTransaction(uniV3RouterAddr, nonSwapCalldata, nil)
	if decoder.Matches(nonSwapTx, uniV3RouterAddr.Hex()) {
		t.Error("Expected decoder to not match transaction with non-swap function signature")
	}
}

func TestUniswapV3Decoder_Protocol(t *testing.T) {
	decoder := NewUniswapV3Decoder()

	if decoder.Protocol() != types.UniswapV3Protocol {
		t.Errorf("Expected protocol %s, got %s", types.UniswapV3Protocol, decoder.Protocol())
	}
}

func TestDecodeExactInputSingle(t *testing.T) {
	decoder := NewUniswapV3Decoder()

	// Create test transaction data for exactInputSingle (0x414bf389)
	// ExactInputSingleParams: tokenIn, tokenOut, fee, recipient, deadline, amountIn, amountOutMinimum, sqrtPriceLimitX96
	calldata, _ := hex.DecodeString("414bf389" + // exactInputSingle signature
		"0000000000000000000000000000000000000000000000000000000000000020" + // offset to params tuple
		"000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48" + // tokenIn (USDC)
		"000000000000000000000000d533a949740bb3306d11b9d5bc00000000000000" + // tokenOut
		"0000000000000000000000000000000000000000000000000000000000000bb8" + // fee (3000)
		"0000000000000000000000001234567890123456789012345678901234567890" + // recipient
		"0000000000000000000000000000000000000000000000000000000065f5e100" + // deadline
		"0000000000000000000000000000000000000000000000000de0b6b3a7640000" + // amountIn (1 ETH)
		"0000000000000000000000000000000000000000000000000000000000000001" + // amountOutMinimum
		"0000000000000000000000000000000000000000000000000000000000000000") // sqrtPriceLimitX96

	routerAddr := common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564")
	tx := testhelpers.CreateTestTransaction(routerAddr, calldata, nil)

	actions, err := decoder.Decode(tx, routerAddr.Hex())
	if err != nil {
		t.Fatalf("Unexpected error decoding transaction: %v", err)
	}

	if len(actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(actions))
	}

	action := actions[0]
	if action.Type != types.SwapAction {
		t.Errorf("Expected action type %s, got %s", types.SwapAction, action.Type)
	}

	if action.Protocol != types.UniswapV3Protocol {
		t.Errorf("Expected protocol %s, got %s", types.UniswapV3Protocol, action.Protocol)
	}
}

func TestDecodeExactOutputSingle(t *testing.T) {
	decoder := NewUniswapV3Decoder()

	// Create test transaction data for exactOutputSingle (0xdb3e2198)
	// ExactOutputSingleParams: tokenIn, tokenOut, fee, recipient, deadline, amountOut, amountInMaximum, sqrtPriceLimitX96
	calldata, _ := hex.DecodeString("db3e2198" + // exactOutputSingle signature
		"0000000000000000000000000000000000000000000000000000000000000020" + // offset to params tuple
		"000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48" + // tokenIn (USDC)
		"000000000000000000000000d533a949740bb3306d11b9d5bc00000000000000" + // tokenOut
		"0000000000000000000000000000000000000000000000000000000000000bb8" + // fee (3000)
		"0000000000000000000000001234567890123456789012345678901234567890" + // recipient
		"0000000000000000000000000000000000000000000000000000000065f5e100" + // deadline
		"0000000000000000000000000000000000000000000000000de0b6b3a7640000" + // amountOut (1 ETH)
		"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" + // amountInMaximum (max uint256)
		"0000000000000000000000000000000000000000000000000000000000000000") // sqrtPriceLimitX96

	routerAddr := common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564")
	tx := testhelpers.CreateTestTransaction(routerAddr, calldata, nil)

	actions, err := decoder.Decode(tx, routerAddr.Hex())
	if err != nil {
		t.Fatalf("Unexpected error decoding transaction: %v", err)
	}

	if len(actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(actions))
	}

	action := actions[0]
	if action.Type != types.SwapAction {
		t.Errorf("Expected action type %s, got %s", types.SwapAction, action.Type)
	}

	if action.Protocol != types.UniswapV3Protocol {
		t.Errorf("Expected protocol %s, got %s", types.UniswapV3Protocol, action.Protocol)
	}
}

func TestDecodePath(t *testing.T) {
	// Test the path decoding functionality
	token0 := common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48") // USDC
	token1 := common.HexToAddress("0xd533a949740bb3306d11b9d5bc00000000000000") // CRV

	// Create a path: token0 -> fee -> token1
	// In Uniswap V3 format: 20 bytes (token0) + 3 bytes (fee) + 20 bytes (token1)
	pathBytes := make([]byte, 0)
	pathBytes = append(pathBytes, token0.Bytes()...)           // 20 bytes for token0
	pathBytes = append(pathBytes, []byte{0x00, 0x0b, 0xb8}...) // 3 bytes for 3000 bps fee
	pathBytes = append(pathBytes, token1.Bytes()...)           // 20 bytes for token1

	decodedPath, err := DecodePath(pathBytes)
	if err != nil {
		t.Fatalf("Unexpected error decoding path: %v", err)
	}

	if len(decodedPath) != 2 {
		t.Fatalf("Expected 2 tokens in path, got %d", len(decodedPath))
	}

	if decodedPath[0] != token0 {
		t.Errorf("Expected first token %s, got %s", token0.Hex(), decodedPath[0].Hex())
	}

	if decodedPath[1] != token1 {
		t.Errorf("Expected second token %s, got %s", token1.Hex(), decodedPath[1].Hex())
	}
}

func TestDecodePathWithFees(t *testing.T) {
	token0 := common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48") // USDC
	token1 := common.HexToAddress("0xd533a949740bb3306d11b9d5bc00000000000000") // CRV

	// Create a path: token0 -> fee -> token1
	pathBytes := make([]byte, 0)
	pathBytes = append(pathBytes, token0.Bytes()...)           // 20 bytes for token0
	pathBytes = append(pathBytes, []byte{0x00, 0x0b, 0xb8}...) // 3 bytes for 3000 bps fee
	pathBytes = append(pathBytes, token1.Bytes()...)           // 20 bytes for token1

	decodedTokens, decodedFees, err := DecodePathWithFees(pathBytes)
	if err != nil {
		t.Fatalf("Unexpected error decoding path with fees: %v", err)
	}

	if len(decodedTokens) != 2 {
		t.Fatalf("Expected 2 tokens in path, got %d", len(decodedTokens))
	}

	if len(decodedFees) != 1 {
		t.Fatalf("Expected 1 fee in path, got %d", len(decodedFees))
	}

	if decodedTokens[0] != token0 {
		t.Errorf("Expected first token %s, got %s", token0.Hex(), decodedTokens[0].Hex())
	}

	if decodedTokens[1] != token1 {
		t.Errorf("Expected second token %s, got %s", token1.Hex(), decodedTokens[1].Hex())
	}

	if decodedFees[0] != 3000 {
		t.Errorf("Expected fee 3000, got %d", decodedFees[0])
	}
}

func TestFormatPath(t *testing.T) {
	token0 := common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48") // USDC
	token1 := common.HexToAddress("0xd533a949740bb3306d11b9d5bc00000000000000") // CRV
	tokens := []common.Address{token0, token1}

	fees := []uint32{3000} // 0.3% fee

	pathBytes, err := FormatPath(tokens, fees)
	if err != nil {
		t.Fatalf("Unexpected error formatting path: %v", err)
	}

	// Decode the formatted path to verify it
	decodedTokens, decodedFees, err := DecodePathWithFees(pathBytes)
	if err != nil {
		t.Fatalf("Unexpected error decoding formatted path: %v", err)
	}

	if len(decodedTokens) != len(tokens) {
		t.Errorf("Expected %d tokens, got %d", len(tokens), len(decodedTokens))
	}

	if len(decodedFees) != len(fees) {
		t.Errorf("Expected %d fees, got %d", len(fees), len(decodedFees))
	}

	for i, token := range decodedTokens {
		if token != tokens[i] {
			t.Errorf("Token at index %d: expected %s, got %s", i, tokens[i].Hex(), token.Hex())
		}
	}

	for i, fee := range decodedFees {
		if fee != fees[i] {
			t.Errorf("Fee at index %d: expected %d, got %d", i, fees[i], fee)
		}
	}
}

func TestDecodePathErrors(t *testing.T) {
	// Test empty path
	_, err := DecodePath([]byte{})
	if err == nil {
		t.Error("Expected error for empty path")
	}

	// Test invalid path (not multiple of 23 bytes)
	invalidPath := make([]byte, 21) // 20 bytes for token + only 1 byte for fee (should be 3)
	_, err = DecodePath(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid path length")
	}

	// Test path with insufficient bytes for fee after last token
	validToken := common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")
	partialPath := append(validToken.Bytes(), []byte{0x00, 0x0b}...) // Only 2 bytes for fee (should be 3)
	_, err = DecodePath(partialPath)
	if err == nil {
		t.Error("Expected error for path with insufficient fee bytes")
	}
}
