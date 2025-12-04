// Package types defines common interfaces used throughout the arbitrum-sequencer-decoder
package types

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

// Decoder is the interface for DEX transaction decoders
type Decoder interface {
	// Matches checks if a transaction should be decoded by this decoder
	Matches(tx *types.Transaction, toAddress string) bool

	// Decode decodes the transaction data and returns the decoded actions
	Decode(tx *types.Transaction, toAddress string) ([]DecodedAction, error)

	// Protocol returns the protocol this decoder handles
	Protocol() ProtocolType
}

// Simulator is the interface for price impact simulators
type Simulator interface {
	// SimulateSwap simulates a swap and returns the expected output
	SimulateSwap(pool *Pool, tokenIn Token, amountIn *big.Int) (*SwapSimulation, error)

	// UpdatePoolState updates the pool state based on a swap
	UpdatePoolState(pool *Pool, action DecodedAction) error
}

// PoolOracle is the interface for tracking pool states
type PoolOracle interface {
	// GetPool returns the current state of a pool
	GetPool(address Address) (*Pool, error)

	// UpdatePools updates pool states with new data
	UpdatePools(pools []*Pool) error

	// GetPoolState returns the detailed state of a pool
	GetPoolState(address Address) (*PoolState, error)

	// UpdatePoolState updates the state of a specific pool
	UpdatePoolState(address Address, state *PoolState) error
}

// ArbitrageEngine is the interface for detecting arbitrage opportunities
type ArbitrageEngine interface {
	// DetectOpportunities detects arbitrage opportunities from decoded transactions
	DetectOpportunities(actions []DecodedAction) ([]ArbitrageOpportunity, error)

	// FindCrossDEXOpportunities finds cross-DEX arbitrage opportunities
	FindCrossDEXOpportunities(pools []*Pool) ([]ArbitrageOpportunity, error)

	// FindTriangleOpportunities finds triangle arbitrage opportunities
	FindTriangleOpportunities(pools []*Pool) ([]ArbitrageOpportunity, error)
}

// TransactionExecutor is the interface for executing arbitrage transactions
type TransactionExecutor interface {
	// Execute executes an arbitrage opportunity
	Execute(opportunity ArbitrageOpportunity) error

	// EstimateGas estimates the gas cost for executing an opportunity
	EstimateGas(opportunity ArbitrageOpportunity) (uint64, error)

	// Validate determines if an opportunity is still valid
	Validate(opportunity ArbitrageOpportunity) (bool, error)
}

// SequencerReader is the interface for reading from the sequencer
type SequencerReader interface {
	// Start starts reading from the sequencer
	Start(ctx context.Context) error

	// Stop stops reading from the sequencer
	Stop() error
}

// TransactionCapturer is the interface for capturing and storing transactions
type TransactionCapturer interface {
	// Capture captures pending transactions and stores them
	Capture(ctx context.Context) error

	// Store stores a transaction to the persistence layer
	Store(tx *Transaction) error
}
