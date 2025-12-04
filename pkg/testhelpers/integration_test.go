package testhelpers

import (
	"math/big"
	"testing"
	"time"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/classifier"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/oracle"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
	"github.com/ethereum/go-ethereum/common"
)

// TestDecoderOracleIntegration tests the integration between decoder and oracle components
func TestDecoderOracleIntegration(t *testing.T) {
	// Create a pool oracle
	poolOracle := oracle.NewPoolOracle()

	// Create a test pool and add it to the oracle
	poolAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	token0 := CreateTestToken(
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
		"WETH",
		"Wrapped Ether",
		18,
	)
	token1 := CreateTestToken(
		common.HexToAddress("0x2222222222222222222222222222222222222222"),
		"USDC",
		"USD Coin",
		6,
	)
	pool := CreateTestPool(poolAddress, token0, token1, types.UniswapV3Pool)

	err := poolOracle.UpdatePools([]*types.Pool{pool})
	if err != nil {
		t.Fatalf("Failed to update pools: %v", err)
	}

	// Create a test pool state and add it to the oracle
	poolState := CreateTestPoolState(big.NewInt(1000000000000000000), big.NewInt(2000000000000))
	err = poolOracle.UpdatePoolState(poolAddress, poolState)
	if err != nil {
		t.Fatalf("Failed to update pool state: %v", err)
	}

	// Verify the pool and state were correctly added
	retrievedPool, err := poolOracle.GetPool(poolAddress)
	if err != nil {
		t.Fatalf("Failed to get pool: %v", err)
	}

	if retrievedPool.Type != types.UniswapV3Pool {
		t.Errorf("Expected pool type %s, got %s", types.UniswapV3Pool, retrievedPool.Type)
	}

	retrievedState, err := poolOracle.GetPoolState(poolAddress)
	if err != nil {
		t.Fatalf("Failed to get pool state: %v", err)
	}

	if retrievedState.Reserves0.Cmp(poolState.Reserves0) != 0 {
		t.Errorf("Expected reserves0 %s, got %s", poolState.Reserves0.String(), retrievedState.Reserves0.String())
	}

	// Test classifier with the pool address
	if !classifier.IsDEXTransaction(poolAddress) {
		// Note: This would be false in this test since we're using a random address
		// In a real scenario, we'd use a known DEX address
		t.Log("Test address is not a known DEX (expected in this test)")
	}
}

// TestEndToEndSimulation demonstrates an end-to-end test of the system
func TestEndToEndSimulation(t *testing.T) {
	// This would test the flow: transaction -> classifier -> decoder -> oracle -> arbitrage engine
	// For now, we'll implement a basic version showing the concept

	// Create oracle
	poolOracle := oracle.NewPoolOracle()

	// Set up test pools
	poolAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	token0 := CreateTestToken(common.HexToAddress("0x1111111111111111111111111111111111111111"), "WETH", "Wrapped Ether", 18)
	token1 := CreateTestToken(common.HexToAddress("0x2222222222222222222222222222222222222222"), "USDC", "USD Coin", 6)
	pool := CreateTestPool(poolAddress, token0, token1, types.UniswapV3Pool)

	// Add to oracle
	err := poolOracle.UpdatePools([]*types.Pool{pool})
	if err != nil {
		t.Fatalf("Failed to update pools: %v", err)
	}

	// Update pool state
	poolState := CreateTestPoolState(big.NewInt(1000000000000000000), big.NewInt(2000000000000))
	err = poolOracle.UpdatePoolState(poolAddress, poolState)
	if err != nil {
		t.Fatalf("Failed to update pool state: %v", err)
	}

	// Simulate transaction processing
	// In a real implementation, this would involve:
	// 1. Reading a transaction from sequencer
	// 2. Classifying it
	// 3. Decoding with appropriate decoder
	// 4. Updating pool states
	// 5. Detecting arbitrage opportunities

	// Verify all test data is accessible
	_, err = poolOracle.GetPool(poolAddress)
	if err != nil {
		t.Errorf("Failed to get pool after setup: %v", err)
	}

	_, err = poolOracle.GetPoolState(poolAddress)
	if err != nil {
		t.Errorf("Failed to get pool state after setup: %v", err)
	}

	// Simulate a time delay and check recent pools
	time.Sleep(10 * time.Millisecond) // Brief delay to test time-based queries
	recentPools := poolOracle.GetRecentPools(1 * time.Second)
	if len(recentPools) == 0 {
		t.Error("Expected to find recent pools")
	}
}
