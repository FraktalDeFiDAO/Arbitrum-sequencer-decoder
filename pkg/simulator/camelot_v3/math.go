// Package camelot_v3 implements price simulation for Camelot V3 pools
package camelot_v3

import (
	"errors"
	"math/big"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
	"github.com/ethereum/go-ethereum/common"
)

// PoolState represents the state of a Camelot V3 pool for simulation
// Camelot V3 is similar to Uniswap V3 but with some variations in fee structures and mechanisms
type PoolState struct {
	Token0          common.Address    // Token0 address
	Token1          common.Address    // Token1 address
	Fee             *big.Int          // Fee tier (as parts per million, e.g., 3000 = 0.3%)
	Reserves0       *big.Int          // Reserve of token0
	Reserves1       *big.Int          // Reserve of token1
	Liquidity       *big.Int          // Current liquidity in the active range
	Price0To1       *big.Float        // Current price of token0 in terms of token1
	Price1To0       *big.Float        // Current price of token1 in terms of token0
	Tick            int               // Current tick (for concentrated liquidity models)
	ActiveLiquidity *big.Int          // Active liquidity at current price
	ExtraData       map[string]string // Protocol-specific data
}

// SwapSimulation represents the result of simulating a swap
type SwapSimulation struct {
	TokenIn           common.Address `json:"token_in"`
	TokenOut          common.Address `json:"token_out"`
	AmountIn          *big.Int       `json:"amount_in"`
	ExpectedAmountOut *big.Int       `json:"expected_amount_out"`
	GasEstimate       uint64         `json:"gas_estimate"`
	PriceImpact       float64        `json:"price_impact"` // Percentage
	Error             error          `json:"error,omitempty"`
}

// SimulateSwap simulates a swap in a Camelot V3 pool
func SimulateSwap(poolState *PoolState, tokenIn common.Address, amountIn *big.Int) (*types.SwapSimulation, error) {
	if poolState.Reserves0.Sign() <= 0 || poolState.Reserves1.Sign() <= 0 {
		return nil, errors.New("invalid pool reserves: must be positive")
	}

	if amountIn.Sign() <= 0 {
		return nil, errors.New("amount in must be positive")
	}

	// Calculate the expected output amount using Camelot V3 algorithm
	// The algorithm will be similar to Uniswap V3 but may have subtle differences
	var expectedAmountOut *big.Int

	if tokenIn == poolState.Token0 {
		// Swapping token0 for token1
		expectedAmountOut = calculateAmountOut(amountIn, poolState.Reserves0, poolState.Reserves1)
	} else if tokenIn == poolState.Token1 {
		// Swapping token1 for token0
		expectedAmountOut = calculateAmountOut(amountIn, poolState.Reserves1, poolState.Reserves0)
	} else {
		return nil, errors.New("token not in pool")
	}

	// Calculate price impact
	priceImpact := calculatePriceImpact(poolState, tokenIn, amountIn)

	simulation := &types.SwapSimulation{
		TokenIn:           types.Token{Address: tokenIn},
		TokenOut:          types.Token{Address: getOtherToken(poolState, tokenIn)},
		AmountIn:          amountIn,
		ExpectedAmountOut: expectedAmountOut,
		GasEstimate:       100000, // Default gas estimate
		PriceImpact:       priceImpact,
	}

	return simulation, nil
}

// calculateAmountOut uses the constant product formula with fees to calculate the output amount
// This is a simplified approach - real Camelot V3 uses concentrated liquidity formulas
func calculateAmountOut(amountIn, reserveIn, reserveOut *big.Int) *big.Int {
	// Apply the fee to the input amount
	// Camelot V3 fees might be different from Uniswap V3
	feeNumerator := big.NewInt(10000) // Assuming 0.3% fee as 30 (this would need to be checked)
	feeDenominator := big.NewInt(10000)

	// Calculate the effective amount after fees
	amountAfterFee := new(big.Int).Mul(amountIn, new(big.Int).Sub(feeDenominator, feeNumerator))
	amountAfterFee.Div(amountAfterFee, feeDenominator)

	// Calculate the numerator and denominator for the constant product formula
	numerator := new(big.Int).Mul(amountAfterFee, reserveOut)
	denominator := new(big.Int).Add(reserveIn, amountAfterFee)

	// Calculate the output amount
	amountOut := new(big.Int).Div(numerator, denominator)

	return amountOut
}

// calculatePriceImpact calculates the price impact of a swap
// This is a non-recursive implementation to prevent stack overflow
func calculatePriceImpact(poolState *PoolState, tokenIn common.Address, amountIn *big.Int) float64 {
	// Calculate the current price before the swap
	var currentPrice *big.Float
	if tokenIn == poolState.Token0 {
		currentPrice = new(big.Float).Quo(
			new(big.Float).SetInt(poolState.Reserves1),
			new(big.Float).SetInt(poolState.Reserves0),
		)
	} else {
		currentPrice = new(big.Float).Quo(
			new(big.Float).SetInt(poolState.Reserves0),
			new(big.Float).SetInt(poolState.Reserves1),
		)
	}

	// Calculate expected output directly without calling SimulateSwap (prevents recursion)
	var expectedAmountOut *big.Int
	if tokenIn == poolState.Token0 {
		expectedAmountOut = calculateAmountOut(amountIn, poolState.Reserves0, poolState.Reserves1)
	} else {
		expectedAmountOut = calculateAmountOut(amountIn, poolState.Reserves1, poolState.Reserves0)
	}

	if expectedAmountOut == nil || expectedAmountOut.Sign() <= 0 {
		return 0.0
	}

	// Calculate the new reserves after the swap
	newReservesIn := new(big.Int)
	newReservesOut := new(big.Int)

	if tokenIn == poolState.Token0 {
		newReservesIn.Add(poolState.Reserves0, amountIn)
		newReservesOut.Sub(poolState.Reserves1, expectedAmountOut)
	} else {
		newReservesIn.Add(poolState.Reserves1, amountIn)
		newReservesOut.Sub(poolState.Reserves0, expectedAmountOut)
	}

	// Validate new reserves are positive
	if newReservesIn.Sign() <= 0 || newReservesOut.Sign() <= 0 {
		return 0.0
	}

	// Calculate the new price after the swap
	newPrice := new(big.Float).Quo(
		new(big.Float).SetInt(newReservesOut),
		new(big.Float).SetInt(newReservesIn),
	)

	// Calculate the price impact percentage
	priceDiff := new(big.Float).Sub(currentPrice, newPrice)
	priceImpact := new(big.Float).Quo(priceDiff, currentPrice)
	priceImpactFloat, _ := priceImpact.Float64()

	return priceImpactFloat * 100 // Convert to percentage
}

// getOtherToken returns the other token in the pool
func getOtherToken(poolState *PoolState, token common.Address) common.Address {
	if token == poolState.Token0 {
		return poolState.Token1
	}
	return poolState.Token0
}

// CalculateLiquidityRequired calculates the liquidity needed for a specific swap
func CalculateLiquidityRequired(poolState *PoolState, tokenIn common.Address, amountIn *big.Int) (*big.Int, error) {
	// This is a simplified implementation - real calculation would be more complex
	// based on the current tick and liquidity distribution in the pool
	if poolState.Reserves0.Sign() <= 0 || poolState.Reserves1.Sign() <= 0 {
		return nil, errors.New("invalid pool reserves")
	}

	// For a basic calculation, we can estimate required liquidity based on reserve ratios
	var requiredLiquidity *big.Int
	if tokenIn == poolState.Token0 {
		// Calculate based on token0 amount
		ratio := new(big.Float).Quo(
			new(big.Float).SetInt(amountIn),
			new(big.Float).SetInt(poolState.Reserves0),
		)

		liquidityFloat := new(big.Float).Mul(ratio, new(big.Float).SetInt(poolState.Liquidity))
		requiredLiquidity, _ = liquidityFloat.Int(nil)
	} else {
		// Calculate based on token1 amount
		ratio := new(big.Float).Quo(
			new(big.Float).SetInt(amountIn),
			new(big.Float).SetInt(poolState.Reserves1),
		)

		liquidityFloat := new(big.Float).Mul(ratio, new(big.Float).SetInt(poolState.Liquidity))
		requiredLiquidity, _ = liquidityFloat.Int(nil)
	}

	return requiredLiquidity, nil
}

// UpdatePoolStateAfterSwap updates the pool state after a swap
func UpdatePoolStateAfterSwap(poolState *PoolState, tokenIn common.Address, amountIn, amountOut *big.Int) error {
	if tokenIn == poolState.Token0 {
		poolState.Reserves0.Add(poolState.Reserves0, amountIn)
		poolState.Reserves1.Sub(poolState.Reserves1, amountOut)
	} else if tokenIn == poolState.Token1 {
		poolState.Reserves1.Add(poolState.Reserves1, amountIn)
		poolState.Reserves0.Sub(poolState.Reserves0, amountOut)
	} else {
		return errors.New("token not in pool")
	}

	// Update prices
	if poolState.Reserves0.Sign() > 0 && poolState.Reserves1.Sign() > 0 {
		poolState.Price0To1 = new(big.Float).Quo(
			new(big.Float).SetInt(poolState.Reserves1),
			new(big.Float).SetInt(poolState.Reserves0),
		)

		poolState.Price1To0 = new(big.Float).Quo(
			new(big.Float).SetInt(poolState.Reserves0),
			new(big.Float).SetInt(poolState.Reserves1),
		)
	}

	return nil
}
