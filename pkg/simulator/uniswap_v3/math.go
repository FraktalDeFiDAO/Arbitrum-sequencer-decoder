// Package uniswap_v3 implements price simulation for Uniswap V3 pools
package uniswap_v3

import (
	"errors"
	"math/big"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
	"github.com/ethereum/go-ethereum/common"
)

// Constants for Uniswap V3 math
var (
	// Q96 represents 2^96 used in sqrtPriceX96
	Q96 = new(big.Int).Exp(big.NewInt(2), big.NewInt(96), nil)

	// Q128 represents 2^128
	Q128 = new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil)

	// Max tick in Uniswap V3
	MaxTick = 887272

	// Min tick in Uniswap V3
	MinTick = -887272

	// TickBase is 1.0001 scaled by 2^128 for precise calculations
	// sqrt(1.0001) = 1.00004999875...
	// We use fixed-point math to avoid floating point errors
	sqrtTickBase *big.Int

	// Fee denominator (1,000,000 = 100%)
	FeeDenominator = big.NewInt(1000000)
)

func init() {
	// Initialize sqrt(1.0001) * 2^64 for tick calculations
	// sqrt(1.0001) ≈ 1.000049998750062
	// Scaled by 2^64: 18447025884307988
	sqrtTickBase = big.NewInt(18447025884307988)
}

// PoolState represents the state of a Uniswap V3 pool for simulation
type PoolState struct {
	SqrtPriceX96 *big.Int         // Square root price scaled by 2^96
	Tick         int              // Current tick
	Liquidity    *big.Int         // Current liquidity in the active tick range
	LiquidityNet map[int]*big.Int // Liquidity changes at each tick
	TickSpacing  int              // Tick spacing for this pool (1, 5, 30, 100)
	Token0       common.Address   // Token0 address
	Token1       common.Address   // Token1 address
	Fee          *big.Int         // Fee tier (as parts per million, e.g., 3000 = 0.3%)
}

// SwapState represents the state during a swap simulation
type SwapState struct {
	AmountSpecifiedRemaining *big.Int
	AmountCalculated         *big.Int
	SqrtPriceX96             *big.Int
	Tick                     int
	Liquidity                *big.Int
}

// SimulateSwap simulates a swap in a Uniswap V3 pool
func SimulateSwap(poolState *PoolState, zeroForOne bool, amountSpecified *big.Int, sqrtPriceLimitX96 *big.Int) (*types.SwapSimulation, error) {
	// Validate inputs
	if poolState.Liquidity.Sign() <= 0 {
		return nil, errors.New("invalid liquidity: must be positive")
	}

	if amountSpecified.Sign() == 0 {
		return nil, errors.New("amount specified must be non-zero")
	}

	// Initialize swap state
	state := &SwapState{
		AmountSpecifiedRemaining: new(big.Int).Set(amountSpecified),
		AmountCalculated:         big.NewInt(0),
		SqrtPriceX96:             new(big.Int).Set(poolState.SqrtPriceX96),
		Tick:                     poolState.Tick,
		Liquidity:                new(big.Int).Set(poolState.Liquidity),
	}

	// The sign of amount specified determines if it's exact input or output
	exactInput := amountSpecified.Sign() > 0

	// Use the swap calculation loop
	for state.AmountSpecifiedRemaining.Sign() != 0 && state.SqrtPriceX96.Cmp(sqrtPriceLimitX96) != 0 {
		// Calculate the next tick where liquidity changes
		nextTick, err := getNextTick(state.Tick, zeroForOne, poolState.LiquidityNet)
		if err != nil {
			return nil, err
		}

		// Calculate the sqrt price at the next tick
		sqrtPriceNextX96, err := getSqrtPriceAtTick(nextTick)
		if err != nil {
			return nil, err
		}

		// Determine if we move to the next tick or stay at the limit
		var stepSqrtPriceX96 *big.Int
		if (zeroForOne && (sqrtPriceNextX96.Cmp(sqrtPriceLimitX96) < 0 || sqrtPriceLimitX96.Cmp(poolState.SqrtPriceX96) < 0)) ||
			(!zeroForOne && (sqrtPriceNextX96.Cmp(sqrtPriceLimitX96) > 0 || sqrtPriceLimitX96.Cmp(poolState.SqrtPriceX96) > 0)) {
			stepSqrtPriceX96 = sqrtPriceLimitX96
		} else {
			stepSqrtPriceX96 = sqrtPriceNextX96
		}

		// Calculate the swap step
		var swapStepSqrtPriceX96, amountIn, amountOut, feeAmount *big.Int
		swapStepSqrtPriceX96, amountIn, amountOut, feeAmount, err = computeSwapStep(
			state.SqrtPriceX96,
			stepSqrtPriceX96,
			state.Liquidity,
			state.AmountSpecifiedRemaining,
			poolState.Fee,
			exactInput,
		)
		if err != nil {
			return nil, err
		}

		// Update the swap state
		state.SqrtPriceX96 = swapStepSqrtPriceX96

		// Add fee amount to amount in if it's exact input
		if exactInput {
			amountIn.Add(amountIn, feeAmount)
		}

		// Update remaining amounts
		state.AmountSpecifiedRemaining.Sub(state.AmountSpecifiedRemaining, amountIn)
		state.AmountCalculated.Sub(state.AmountCalculated, amountOut)

		// Update liquidity if we crossed a tick
		if state.SqrtPriceX96.Cmp(sqrtPriceNextX96) == 0 {
			liquidityNet, exists := poolState.LiquidityNet[nextTick]
			if !exists {
				// This should not happen in a valid pool state, but let's handle it gracefully
				liquidityNet = big.NewInt(0)
			}

			if zeroForOne {
				state.Liquidity.Sub(state.Liquidity, liquidityNet)
			} else {
				state.Liquidity.Add(state.Liquidity, liquidityNet)
			}

			state.Tick = nextTick
		} else if state.SqrtPriceX96.Cmp(stepSqrtPriceX96) != 0 {
			// If the price didn't reach the next tick, calculate the new tick
			newTick, err := getTickAtSqrtPrice(state.SqrtPriceX96)
			if err != nil {
				return nil, err
			}
			state.Tick = newTick
		}
	}

	// Create simulation result
	simulation := &types.SwapSimulation{
		AmountIn:          new(big.Int).Sub(amountSpecified, state.AmountSpecifiedRemaining),
		ExpectedAmountOut: state.AmountCalculated,
		// Price impact calculation would require more complex logic
		PriceImpact: 0.0, // Placeholder
	}

	return simulation, nil
}

// computeSwapStep calculates a single step of the swap
func computeSwapStep(sqrtPX96, sqrtPNextX96, liquidity, amountRemaining *big.Int, fee *big.Int, exactInput bool) (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	var sqrtPriceX96, amountIn, amountOut, feeAmount *big.Int
	var err error

	// Calculate the square root price for this step
	if exactInput {
		amountIn, err = getAmount0Delta(sqrtPX96, sqrtPNextX96, liquidity, true)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		if amountRemaining.Cmp(amountIn) < 0 {
			// Not enough remaining amount, adjust price
			amountIn = amountRemaining
			feeAmount = getFeeAmount(amountIn, fee)
			amountInWithFee := new(big.Int).Add(amountIn, feeAmount)

			sqrtPriceX96, err = getSqrtPriceX96WithDelta(sqrtPX96, liquidity, amountInWithFee, true)
			if err != nil {
				return nil, nil, nil, nil, err
			}

			amountOut, err = getAmount1Delta(sqrtPX96, sqrtPriceX96, liquidity, false)
			if err != nil {
				return nil, nil, nil, nil, err
			}
		} else {
			feeAmount = getFeeAmount(amountIn, fee)
			sqrtPriceX96 = sqrtPNextX96
			amountOut, err = getAmount1Delta(sqrtPX96, sqrtPNextX96, liquidity, false)
			if err != nil {
				return nil, nil, nil, nil, err
			}
		}
	} else {
		amountOut, err = getAmount1Delta(sqrtPX96, sqrtPNextX96, liquidity, false)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		if amountRemaining.Cmp(amountOut) < 0 {
			// Not enough remaining amount, adjust price
			amountOut = amountRemaining
			sqrtPriceX96, err = getSqrtPriceX96WithDelta(sqrtPX96, liquidity, amountOut, false)
			if err != nil {
				return nil, nil, nil, nil, err
			}

			amountIn, err = getAmount0Delta(sqrtPriceX96, sqrtPX96, liquidity, true)
			if err != nil {
				return nil, nil, nil, nil, err
			}

			feeAmount = getFeeAmount(amountIn, fee)
		} else {
			sqrtPriceX96 = sqrtPNextX96
			amountIn, err = getAmount0Delta(sqrtPX96, sqrtPNextX96, liquidity, true)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			feeAmount = getFeeAmount(amountIn, fee)
		}
	}

	return sqrtPriceX96, amountIn, amountOut, feeAmount, nil
}

// getAmount0Delta calculates the amount of token0 for a given price change
func getAmount0Delta(sqrtA, sqrtB, liquidity *big.Int, roundUp bool) (*big.Int, error) {
	if sqrtA.Cmp(sqrtB) > 0 {
		sqrtA, sqrtB = sqrtB, sqrtA
	}

	numerator := new(big.Int).Mul(liquidity, new(big.Int).Sub(sqrtB, sqrtA))
	denominator := new(big.Int).Mul(sqrtA, sqrtB)

	if denominator.Sign() == 0 {
		return nil, errors.New("division by zero in getAmount0Delta")
	}

	result := new(big.Int).Quo(numerator, denominator)
	if roundUp && new(big.Int).Rem(numerator, denominator).Sign() != 0 {
		result.Add(result, big.NewInt(1))
	}

	return result, nil
}

// getAmount1Delta calculates the amount of token1 for a given price change
func getAmount1Delta(sqrtA, sqrtB, liquidity *big.Int, roundUp bool) (*big.Int, error) {
	if sqrtA.Cmp(sqrtB) > 0 {
		sqrtA, sqrtB = sqrtB, sqrtA
	}

	result := new(big.Int).Mul(liquidity, new(big.Int).Sub(sqrtB, sqrtA))
	result.Div(result, Q96)

	if roundUp && new(big.Int).Rem(new(big.Int).Mul(liquidity, new(big.Int).Sub(sqrtB, sqrtA)), Q96).Sign() != 0 {
		result.Add(result, big.NewInt(1))
	}

	return result, nil
}

// getFeeAmount calculates the fee amount from the input amount
func getFeeAmount(amount, fee *big.Int) *big.Int {
	// Fee is typically in hundredths of basis points (e.g., 3000 = 0.3%)
	product := new(big.Int).Mul(amount, fee)
	// Divide by 1,000,000 to get the fee (0.3% = 3000/1,000,000)
	return new(big.Int).Quo(product, big.NewInt(1000000))
}

// getNextTick finds the next tick in the specified direction
func getNextTick(currentTick int, zeroForOne bool, liquidityNet map[int]*big.Int) (int, error) {
	// In a real implementation, this would search through initialized ticks
	// For this simplified implementation, we'll return the next logical tick
	// based on common tick spacings (1, 5, 30, 100)

	if zeroForOne {
		// For zeroForOne swaps, we move to lower ticks
		for tick := currentTick; tick >= MinTick; tick-- {
			if _, exists := liquidityNet[tick]; exists {
				return tick, nil
			}
		}
	} else {
		// For oneForZero swaps, we move to higher ticks
		for tick := currentTick + 1; tick <= MaxTick; tick++ {
			if _, exists := liquidityNet[tick]; exists {
				return tick, nil
			}
		}
	}

	// If no initialized tick is found, return an error
	return 0, errors.New("no next initialized tick found")
}

// getSqrtPriceAtTick calculates the square root price at a given tick
// Formula: sqrtPrice = sqrt(1.0001^tick) * 2^96 = 1.0001^(tick/2) * 2^96
func getSqrtPriceAtTick(tick int) (*big.Int, error) {
	if tick < MinTick || tick > MaxTick {
		return nil, errors.New("tick out of bounds")
	}

	absTick := tick
	if absTick < 0 {
		absTick = -absTick
	}

	// Use precomputed magic numbers from Uniswap V3
	// These represent sqrt(1.0001^(2^i)) * 2^128 for i = 0..19
	var ratio *big.Int

	// Start with 2^128
	ratio = new(big.Int).Set(Q128)

	// Multiply by precomputed ratios based on tick bits
	if absTick&0x1 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0xfffcb933bd6fad37))
	}
	if absTick&0x2 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0xfff97272373d413d))
	}
	if absTick&0x4 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0xfff2e50f5f656932))
	}
	if absTick&0x8 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0xffe5caca7e10e4e6))
	}
	if absTick&0x10 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0xffcb9843d60f6159))
	}
	if absTick&0x20 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0xff973b41fa98c081))
	}
	if absTick&0x40 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0xff2ea16466c96a3a))
	}
	if absTick&0x80 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0xfe5dee046a99a2a0))
	}
	if absTick&0x100 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0xfcbe86c7900a88ae))
	}
	if absTick&0x200 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0xf987a7253ac413e0))
	}
	if absTick&0x400 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0xf3392b0822b70005))
	}
	if absTick&0x800 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0xe7159475a2c29b7d))
	}
	if absTick&0x1000 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0xd097f3bdfd2022b8))
	}
	if absTick&0x2000 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0xa9f746462d870fdf))
	}
	if absTick&0x4000 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0x70d869a156d2a1b2))
	}
	if absTick&0x8000 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0x31be135f97d08fd9))
	}
	if absTick&0x10000 != 0 {
		ratio = mulShift(ratio, new(big.Int).SetUint64(0x9aa508b5b7a84e1c))
	}
	if absTick&0x20000 != 0 {
		// Value too large for uint64, use SetString
		val, _ := new(big.Int).SetString("5d6af8dedb81196699c329225ee604", 16)
		ratio = mulShift(ratio, val)
	}
	if absTick&0x40000 != 0 {
		val, _ := new(big.Int).SetString("2216e584f5fa1ea926041bedfe98", 16)
		ratio = mulShift(ratio, val)
	}
	if absTick&0x80000 != 0 {
		val, _ := new(big.Int).SetString("48a170391f7dc42444e8fa2", 16)
		ratio = mulShift(ratio, val)
	}

	// If tick is positive, we need to invert the ratio
	if tick > 0 {
		maxUint256 := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))
		ratio = new(big.Int).Div(maxUint256, ratio)
	}

	// Convert from Q128.128 to Q64.96
	// Shift right by 32 bits (128 - 96 = 32)
	result := new(big.Int).Rsh(ratio, 32)

	return result, nil
}

// mulShift multiplies two numbers and shifts right by 64 bits
func mulShift(a, b *big.Int) *big.Int {
	product := new(big.Int).Mul(a, b)
	return new(big.Int).Rsh(product, 64)
}

// getTickAtSqrtPrice calculates the tick at a given square root price
// This is the inverse of getSqrtPriceAtTick
func getTickAtSqrtPrice(sqrtPriceX96 *big.Int) (int, error) {
	if sqrtPriceX96 == nil || sqrtPriceX96.Sign() <= 0 {
		return 0, errors.New("invalid sqrt price")
	}

	// Calculate log base sqrt(1.0001) of (sqrtPriceX96 / 2^96)
	// tick = log_{sqrt(1.0001)}(sqrtPriceX96 / 2^96)
	// tick = 2 * log_{1.0001}(sqrtPriceX96 / 2^96)
	// tick = 2 * ln(sqrtPriceX96 / 2^96) / ln(1.0001)

	// Use binary search for efficiency
	low := MinTick
	high := MaxTick

	for low < high {
		mid := (low + high + 1) / 2
		sqrtPriceAtMid, err := getSqrtPriceAtTick(mid)
		if err != nil {
			return 0, err
		}

		if sqrtPriceAtMid.Cmp(sqrtPriceX96) <= 0 {
			low = mid
		} else {
			high = mid - 1
		}
	}

	return low, nil
}

// getSqrtPriceX96WithDelta calculates the square root price given a price change
func getSqrtPriceX96WithDelta(sqrtPX96, liquidity, amountDelta *big.Int, zeroForOne bool) (*big.Int, error) {
	if zeroForOne {
		// Moving from token0 to token1 (price of token0 decreases)
		price0Delta := new(big.Int).Mul(amountDelta, Q96)
		price0Delta.Div(price0Delta, liquidity)

		newSqrtPX96 := new(big.Int).Sub(sqrtPX96, price0Delta)
		if newSqrtPX96.Sign() <= 0 {
			return nil, errors.New("sqrtPrice would be zero or negative")
		}
		return newSqrtPX96, nil
	} else {
		// Moving from token1 to token0 (price of token0 increases)
		price1Delta := new(big.Int).Mul(amountDelta, Q96)
		price1Delta.Div(price1Delta, liquidity)

		newSqrtPX96 := new(big.Int).Add(sqrtPX96, price1Delta)
		return newSqrtPX96, nil
	}
}

// CalculatePriceImpact calculates the price impact of a swap
func CalculatePriceImpact(poolState *PoolState, amountIn *big.Int, tokenIn common.Address) (float64, error) {
	// Simulate the swap
	_, err := SimulateSwap(poolState, tokenIn == poolState.Token0, amountIn, big.NewInt(0))
	if err != nil {
		return 0, err
	}

	// Calculate price impact as percentage
	// This is a simplified calculation for now
	return 0.0, nil // Placeholder
}

// calculateCurrentPrice calculates the current price from sqrtPriceX96
func calculateCurrentPrice(sqrtPriceX96 *big.Int) float64 {
	// Price0To1 = (sqrtPriceX96 / 2^96)^2
	sqrtPriceFloat := new(big.Float).SetInt(sqrtPriceX96)
	scaleFactorFloat := new(big.Float).SetInt(Q96)
	priceRatio := new(big.Float).Quo(sqrtPriceFloat, scaleFactorFloat)

	// Square to get the actual price
	priceFloat := new(big.Float).Mul(priceRatio, priceRatio)
	result, _ := priceFloat.Float64()
	return result
}
