package oracle

import (
	"math/big"
	"testing"
	"time"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
	"github.com/ethereum/go-ethereum/common"
)

func TestNewPoolOracle(t *testing.T) {
	oracle := NewPoolOracle()

	if oracle.pools == nil {
		t.Error("Expected pools map to be initialized")
	}

	if oracle.states == nil {
		t.Error("Expected states map to be initialized")
	}
}

func TestGetPool(t *testing.T) {
	oracle := NewPoolOracle()

	address := common.HexToAddress("0x1234567890123456789012345678901234567890")
	pool := &types.Pool{
		Address: address,
		Type:    types.UniswapV3Pool,
	}

	// Test getting non-existent pool
	_, err := oracle.GetPool(address)
	if err == nil {
		t.Error("Expected error when getting non-existent pool")
	}
	if err != types.ErrPoolNotFound {
		t.Errorf("Expected ErrPoolNotFound, got %v", err)
	}

	// Add pool and test getting it
	err = oracle.UpdatePools([]*types.Pool{pool})
	if err != nil {
		t.Errorf("Unexpected error updating pools: %v", err)
	}

	retrievedPool, err := oracle.GetPool(address)
	if err != nil {
		t.Errorf("Unexpected error getting pool: %v", err)
	}

	if retrievedPool.Address != pool.Address {
		t.Errorf("Expected address %s, got %s", pool.Address.Hex(), retrievedPool.Address.Hex())
	}

	if retrievedPool.Type != pool.Type {
		t.Errorf("Expected type %s, got %s", pool.Type, retrievedPool.Type)
	}
}

func TestUpdatePools(t *testing.T) {
	oracle := NewPoolOracle()

	pool1 := &types.Pool{
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Type:    types.UniswapV3Pool,
	}

	pool2 := &types.Pool{
		Address: common.HexToAddress("0x0987654321098765432109876543210987654321"),
		Type:    types.CamelotV3Pool,
	}

	pools := []*types.Pool{pool1, pool2}
	err := oracle.UpdatePools(pools)
	if err != nil {
		t.Errorf("Unexpected error updating pools: %v", err)
	}

	// Check if both pools were added
	_, err = oracle.GetPool(pool1.Address)
	if err != nil {
		t.Errorf("Pool1 not found after update: %v", err)
	}

	_, err = oracle.GetPool(pool2.Address)
	if err != nil {
		t.Errorf("Pool2 not found after update: %v", err)
	}
}

func TestGetPoolState(t *testing.T) {
	oracle := NewPoolOracle()

	address := common.HexToAddress("0x1234567890123456789012345678901234567890")
	state := &types.PoolState{
		Reserves0: big.NewInt(1000),
		Reserves1: big.NewInt(2000),
	}

	// Test getting non-existent pool state
	_, err := oracle.GetPoolState(address)
	if err == nil {
		t.Error("Expected error when getting non-existent pool state")
	}
	if err != types.ErrPoolNotFound {
		t.Errorf("Expected ErrPoolNotFound, got %v", err)
	}

	// Add state and test getting it
	err = oracle.UpdatePoolState(address, state)
	if err != nil {
		t.Errorf("Unexpected error updating pool state: %v", err)
	}

	retrievedState, err := oracle.GetPoolState(address)
	if err != nil {
		t.Errorf("Unexpected error getting pool state: %v", err)
	}

	if retrievedState.Reserves0.Cmp(state.Reserves0) != 0 {
		t.Errorf("Expected reserves0 %s, got %s", state.Reserves0.String(), retrievedState.Reserves0.String())
	}

	if retrievedState.Reserves1.Cmp(state.Reserves1) != 0 {
		t.Errorf("Expected reserves1 %s, got %s", state.Reserves1.String(), retrievedState.Reserves1.String())
	}
}

func TestUpdatePoolState(t *testing.T) {
	oracle := NewPoolOracle()

	address := common.HexToAddress("0x1234567890123456789012345678901234567890")
	state := &types.PoolState{
		Reserves0: big.NewInt(1000),
		Reserves1: big.NewInt(2000),
	}

	// Test with nil state
	err := oracle.UpdatePoolState(address, nil)
	if err == nil {
		t.Error("Expected error when updating with nil state")
	}

	// Test successful update
	err = oracle.UpdatePoolState(address, state)
	if err != nil {
		t.Errorf("Unexpected error updating pool state: %v", err)
	}

	// Verify the state was updated and timestamp was set recently
	retrievedState, err := oracle.GetPoolState(address)
	if err != nil {
		t.Errorf("Error retrieving updated state: %v", err)
	}

	if retrievedState.LastUpdated.Before(time.Now().Add(-1 * time.Minute)) {
		t.Error("Expected LastUpdated to be set to current time")
	}
}

func TestRemovePool(t *testing.T) {
	oracle := NewPoolOracle()

	address := common.HexToAddress("0x1234567890123456789012345678901234567890")
	pool := &types.Pool{
		Address: address,
		Type:    types.UniswapV3Pool,
	}
	state := &types.PoolState{
		Reserves0: big.NewInt(1000),
		Reserves1: big.NewInt(2000),
	}

	// Add pool and state
	oracle.UpdatePools([]*types.Pool{pool})
	oracle.UpdatePoolState(address, state)

	// Verify they were added
	if !oracle.IsPoolTracked(address) {
		t.Error("Expected pool to be tracked")
	}

	// Remove the pool
	err := oracle.RemovePool(address)
	if err != nil {
		t.Errorf("Unexpected error removing pool: %v", err)
	}

	// Verify they were removed
	if oracle.IsPoolTracked(address) {
		t.Error("Expected pool to not be tracked after removal")
	}

	_, err = oracle.GetPool(address)
	if err == nil {
		t.Error("Expected error when getting removed pool")
	}

	_, err = oracle.GetPoolState(address)
	if err == nil {
		t.Error("Expected error when getting removed pool state")
	}
}

func TestGetAllPools(t *testing.T) {
	oracle := NewPoolOracle()

	pool1 := &types.Pool{
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Type:    types.UniswapV3Pool,
	}

	pool2 := &types.Pool{
		Address: common.HexToAddress("0x0987654321098765432109876543210987654321"),
		Type:    types.CamelotV3Pool,
	}

	oracle.UpdatePools([]*types.Pool{pool1, pool2})

	allPools := oracle.GetAllPools()
	if len(allPools) != 2 {
		t.Errorf("Expected 2 pools, got %d", len(allPools))
	}

	// Verify both pools are present
	foundPool1 := false
	foundPool2 := false
	for _, pool := range allPools {
		if pool.Address == pool1.Address {
			foundPool1 = true
		}
		if pool.Address == pool2.Address {
			foundPool2 = true
		}
	}

	if !foundPool1 || !foundPool2 {
		t.Error("Expected both pools to be in GetAllPools result")
	}
}

func TestIsPoolTracked(t *testing.T) {
	oracle := NewPoolOracle()

	address := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Initially not tracked
	if oracle.IsPoolTracked(address) {
		t.Error("Expected pool to not be tracked initially")
	}

	// Add pool
	pool := &types.Pool{
		Address: address,
		Type:    types.UniswapV3Pool,
	}
	oracle.UpdatePools([]*types.Pool{pool})

	// Now should be tracked
	if !oracle.IsPoolTracked(address) {
		t.Error("Expected pool to be tracked after adding")
	}
}

func TestConcurrentAccess(t *testing.T) {
	oracle := NewPoolOracle()
	address := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Run multiple goroutines to test concurrent access
	done := make(chan bool)

	// Goroutine for updating pools
	go func() {
		for i := 0; i < 100; i++ {
			pool := &types.Pool{
				Address: address,
				Type:    types.UniswapV3Pool,
			}
			oracle.UpdatePools([]*types.Pool{pool})
		}
		done <- true
	}()

	// Goroutine for getting pools
	go func() {
		for i := 0; i < 100; i++ {
			oracle.GetPool(address)
		}
		done <- true
	}()

	// Goroutine for updating pool states
	go func() {
		for i := 0; i < 100; i++ {
			state := &types.PoolState{
				Reserves0: big.NewInt(int64(i)),
				Reserves1: big.NewInt(int64(i * 2)),
			}
			oracle.UpdatePoolState(address, state)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	<-done
	<-done
	<-done
}
