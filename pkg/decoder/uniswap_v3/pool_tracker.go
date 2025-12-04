// Package uniswap_v3 provides Uniswap V3 specific pool state tracking
package uniswap_v3

import (
	"math/big"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/oracle"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
	"github.com/ethereum/go-ethereum/common"
)

// PoolExtension holds Uniswap V3-specific pool data
type PoolExtension struct {
	Liquidity       *big.Int         `json:"liquidity"`
	SqrtPriceX96    *big.Int         `json:"sqrt_price_x96"`
	Tick            int              `json:"tick"`
	TickSpacing     int              `json:"tick_spacing"`
	LiquidityNet    map[int]*big.Int `json:"liquidity_net,omitempty"` // Liquidity changes at each tick
	ActiveLiquidity *big.Int         `json:"active_liquidity"`
}

// UniswapV3PoolTracker provides Uniswap V3 specific pool state tracking
type UniswapV3PoolTracker struct {
	oracle *oracle.PoolOracle
}

// NewUniswapV3PoolTracker creates a new Uniswap V3 pool tracker
func NewUniswapV3PoolTracker(oracle *oracle.PoolOracle) *UniswapV3PoolTracker {
	return &UniswapV3PoolTracker{
		oracle: oracle,
	}
}

// UpdatePoolState updates the Uniswap V3 specific state for a pool
func (t *UniswapV3PoolTracker) UpdatePoolState(address common.Address, extension *PoolExtension) error {
	// First get the existing state
	existingState, err := t.oracle.GetPoolState(address)
	if err != nil {
		// If pool doesn't exist, create a new one
		if err == types.ErrPoolNotFound {
			existingState = &types.PoolState{}
		} else {
			return err
		}
	}

	// Update Uniswap V3 specific fields in the extra data
	if existingState.ExtraData == nil {
		existingState.ExtraData = make(map[string]string)
	}

	// Update the main pool state with relevant information
	existingState.Reserves0 = extension.SqrtPriceX96 // Placeholder - in real implementation we'd calculate actual reserves
	existingState.Reserves1 = extension.Liquidity
	existingState.Tick = extension.Tick
	existingState.ActiveLiquidity = extension.ActiveLiquidity

	// Store Uniswap V3 specific data as JSON in extra data or we can extend the structure
	// For now, we'll update the generic state and assume the specific data is tracked elsewhere

	return t.oracle.UpdatePoolState(address, existingState)
}

// GetPoolState retrieves the Uniswap V3 specific state for a pool
func (t *UniswapV3PoolTracker) GetPoolState(address common.Address) (*PoolExtension, error) {
	state, err := t.oracle.GetPoolState(address)
	if err != nil {
		return nil, err
	}

	// In a real implementation, we would parse the extra data or maintain a separate mapping
	// For simplicity, we'll return a basic extension based on the available data
	extension := &PoolExtension{
		Liquidity:       state.Reserves1,
		Tick:            state.Tick,
		ActiveLiquidity: state.ActiveLiquidity,
		LiquidityNet:    make(map[int]*big.Int), // Empty map as placeholder
	}

	return extension, nil
}

// UpdatePoolStateFromSwap updates the pool state based on a swap
func (t *UniswapV3PoolTracker) UpdatePoolStateFromSwap(address common.Address, amount0In, amount1In, amount0Out, amount1Out *big.Int) error {
	currentState, err := t.oracle.GetPoolState(address)
	if err != nil {
		return err
	}

	// Calculate the new reserves based on the swap
	// This is simplified - in reality, Uniswap V3 uses a complex tick math system
	newReserves0 := new(big.Int).Add(currentState.Reserves0, amount0In)
	newReserves0.Sub(newReserves0, amount0Out)

	newReserves1 := new(big.Int).Add(currentState.Reserves1, amount1In)
	newReserves1.Sub(newReserves1, amount1Out)

	// Update the state
	currentState.Reserves0 = newReserves0
	currentState.Reserves1 = newReserves1

	// Recalculate prices
	if currentState.Reserves0.Sign() > 0 {
		price0To1 := new(big.Float).Quo(
			new(big.Float).SetInt(currentState.Reserves1),
			new(big.Float).SetInt(currentState.Reserves0),
		)
		currentState.Price0To1, _ = price0To1.Float64()
	}

	if currentState.Reserves1.Sign() > 0 {
		price1To0 := new(big.Float).Quo(
			new(big.Float).SetInt(currentState.Reserves0),
			new(big.Float).SetInt(currentState.Reserves1),
		)
		currentState.Price1To0, _ = price1To0.Float64()
	}

	return t.oracle.UpdatePoolState(address, currentState)
}

// GetTickAtPrice returns the tick corresponding to a given price
func (t *UniswapV3PoolTracker) GetTickAtPrice(sqrtPriceX96 *big.Int) int {
	// Uniswap V3 uses complex tick math
	// The tick is calculated using log base 1.0001 of the price
	// This is a simplified implementation - in practice, this requires precise mathematical functions
	return 0 // Placeholder
}
