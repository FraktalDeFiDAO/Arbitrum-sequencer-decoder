// Package testhelpers provides utilities and helpers for testing the arbitrum-sequencer-decoder
package testhelpers

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"time"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// CreateTestTransaction creates a mock Ethereum transaction for testing
func CreateTestTransaction(to common.Address, data []byte, value *big.Int) *ethtypes.Transaction {
	if value == nil {
		value = big.NewInt(0)
	}

	// Create a mock transaction - in real implementation this would need proper signing
	tx := ethtypes.NewTx(&ethtypes.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1000000000), // 1 gwei
		Gas:      200000,
		To:       &to,
		Value:    value,
		Data:     data,
	})

	return tx
}

// CreateTestPool creates a test pool with specified parameters
func CreateTestPool(address common.Address, token0, token1 types.Token, poolType types.PoolType) *types.Pool {
	return &types.Pool{
		Address:     address,
		Type:        poolType,
		Token0:      token0,
		Token1:      token1,
		Fee:         big.NewInt(3000),      // Default 0.3% fee for many DEXes
		Liquidity:   big.NewFloat(1000000), // 1M in liquidity
		LastUpdated: time.Now(),
	}
}

// CreateTestToken creates a test token with specified parameters
func CreateTestToken(address common.Address, symbol, name string, decimals uint8) types.Token {
	return types.Token{
		Address:  address,
		Symbol:   symbol,
		Name:     name,
		Decimals: decimals,
	}
}

// CreateTestPoolState creates a test pool state with specified parameters
func CreateTestPoolState(reserves0, reserves1 *big.Int) *types.PoolState {
	if reserves0 == nil {
		reserves0 = big.NewInt(1000000)
	}
	if reserves1 == nil {
		reserves1 = big.NewInt(1000000)
	}

	// Calculate prices based on reserves
	var price0To1, price1To0 float64
	if reserves0.Sign() > 0 {
		result := new(big.Float).Quo(
			new(big.Float).SetInt(reserves1),
			new(big.Float).SetInt(reserves0),
		)
		price0To1, _ = result.Float64()
	}
	if reserves1.Sign() > 0 {
		result := new(big.Float).Quo(
			new(big.Float).SetInt(reserves0),
			new(big.Float).SetInt(reserves1),
		)
		price1To0, _ = result.Float64()
	}

	return &types.PoolState{
		Reserves0:   reserves0,
		Reserves1:   reserves1,
		Price0To1:   price0To1,
		Price1To0:   price1To0,
		LastUpdated: time.Now(),
	}
}

// ReadJSONLFile reads a JSONL (JSON Lines) file and returns the parsed objects
func ReadJSONLFile(filename string, objType interface{}) ([]interface{}, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var results []interface{}
	decoder := json.NewDecoder(file)

	for {
		// Create a new instance of the object type for each line
		obj := objType
		if err := decoder.Decode(&obj); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode line: %w", err)
		}

		results = append(results, obj)
	}

	return results, nil
}

// CreateTestDecodedAction creates a test decoded action for testing
func CreateTestDecodedAction(actionType types.ActionType, protocol types.ProtocolType, pool common.Address,
	tokenIn, tokenOut types.Token, amountIn, amountOut *big.Int) types.DecodedAction {
	return types.DecodedAction{
		Type:      actionType,
		Protocol:  protocol,
		Pool:      pool,
		TokenIn:   tokenIn,
		TokenOut:  tokenOut,
		AmountIn:  amountIn,
		AmountOut: amountOut,
	}
}

// CreateTestArbitrageOpportunity creates a test arbitrage opportunity for testing
func CreateTestArbitrageOpportunity(oppType types.ArbitrageType, profit *big.Int, profitToken types.Token) types.ArbitrageOpportunity {
	return types.ArbitrageOpportunity{
		Type:        oppType,
		Profit:      profit,
		ProfitToken: profitToken,
		Probability: 0.8,                              // 80% probability
		RiskFactor:  0.2,                              // 20% risk factor
		Expiration:  time.Now().Add(30 * time.Second), // Valid for 30 seconds
	}
}
