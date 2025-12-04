package classifier

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestIsDEXTransaction(t *testing.T) {
	// Test with a known DEX address
	uniswapV3Addr := common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564")
	if !IsDEXTransaction(uniswapV3Addr) {
		t.Errorf("Expected %s to be a DEX address", uniswapV3Addr.Hex())
	}

	// Test with a random address
	randomAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	if IsDEXTransaction(randomAddr) {
		t.Errorf("Expected %s to not be a DEX address", randomAddr.Hex())
	}
}

func TestGetDEXProtocol(t *testing.T) {
	uniswapV3Addr := common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564")
	expectedProtocol := "UniswapV3SwapRouter"
	actualProtocol := GetDEXProtocol(uniswapV3Addr)
	if actualProtocol != expectedProtocol {
		t.Errorf("Expected protocol %s, got %s", expectedProtocol, actualProtocol)
	}

	randomAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	if GetDEXProtocol(randomAddr) != "" {
		t.Errorf("Expected empty protocol for unknown address")
	}
}

func TestGetFunctionType(t *testing.T) {
	// Test with Uniswap V3 exactInput signature (0xc04b8d59)
	calldata := common.Hex2Bytes("c04b8d5912345678") // 4 bytes signature + extra data
	expectedFunctionType := "UniswapV3: exactInput"
	actualFunctionType := GetFunctionType(calldata)
	if actualFunctionType != expectedFunctionType {
		t.Errorf("Expected function type %s, got %s", expectedFunctionType, actualFunctionType)
	}

	// Test with short calldata (should return empty)
	shortCalldata := common.Hex2Bytes("1234") // Less than 4 bytes
	if GetFunctionType(shortCalldata) != "" {
		t.Errorf("Expected empty function type for short calldata")
	}

	// Test with unknown signature
	unknownCalldata := common.Hex2Bytes("1234567812345678")
	if GetFunctionType(unknownCalldata) != "" {
		t.Errorf("Expected empty function type for unknown signature")
	}
}

func TestClassifyTransaction(t *testing.T) {
	// Test with known DEX and known function
	uniswapV3Addr := common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564")
	calldata := common.Hex2Bytes("c04b8d5912345678") // exactInput
	expectedProtocol := "UniswapV3SwapRouter"
	expectedFunctionType := "exactInput"

	protocol, functionType, isDEX := ClassifyTransaction(uniswapV3Addr, calldata)
	if !isDEX {
		t.Errorf("Expected transaction to be classified as DEX")
	}
	if protocol != expectedProtocol {
		t.Errorf("Expected protocol %s, got %s", expectedProtocol, protocol)
	}
	if functionType != expectedFunctionType {
		t.Errorf("Expected function type %s, got %s", expectedFunctionType, functionType)
	}

	// Test with known DEX but unknown function
	unknownFunctionCalldata := common.Hex2Bytes("1234567812345678")
	protocol, functionType, isDEX = ClassifyTransaction(uniswapV3Addr, unknownFunctionCalldata)
	if !isDEX {
		t.Errorf("Expected transaction to be classified as DEX")
	}
	if protocol != expectedProtocol {
		t.Errorf("Expected protocol %s, got %s", expectedProtocol, protocol)
	}
	if functionType != "GenericDEX" {
		t.Errorf("Expected function type GenericDEX for unknown function, got %s", functionType)
	}

	// Test with unknown address
	randomAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	protocol, functionType, isDEX = ClassifyTransaction(randomAddr, calldata)
	if isDEX {
		t.Errorf("Expected transaction to NOT be classified as DEX")
	}
	if protocol != "" {
		t.Errorf("Expected empty protocol for unknown address, got %s", protocol)
	}
	if functionType != "" {
		t.Errorf("Expected empty function type for unknown address, got %s", functionType)
	}
}
