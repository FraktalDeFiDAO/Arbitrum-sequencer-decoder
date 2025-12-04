// Package types defines common types and interfaces used throughout the arbitrum-sequencer-decoder
package types

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Address represents an Ethereum address
type Address = common.Address

// Hash represents an Ethereum hash
type Hash = common.Hash

// Token represents a token with its properties
type Token struct {
	Address  Address `json:"address"`
	Symbol   string  `json:"symbol"`
	Name     string  `json:"name"`
	Decimals uint8   `json:"decimals"`
}

// Pool represents a DEX pool with its properties
type Pool struct {
	Address     Address           `json:"address"`
	Type        PoolType          `json:"type"`
	Token0      Token             `json:"token0"`
	Token1      Token             `json:"token1"`
	Token2      *Token            `json:"token2,omitempty"`     // For 3-token pools like some Curve pools
	Fee         *big.Int          `json:"fee"`                  // Fee as a percentage scaled by 1e18, or specific fee structure for different protocols
	ExtraData   map[string]string `json:"extra_data,omitempty"` // Protocol-specific data
	Liquidity   *big.Float        `json:"liquidity"`            // Estimated liquidity in USD
	LastUpdated time.Time         `json:"last_updated"`
}

// PoolType represents the type of DEX pool
type PoolType string

const (
	UniswapV2Pool    PoolType = "uniswap_v2"
	UniswapV3Pool    PoolType = "uniswap_v3"
	CamelotV3Pool    PoolType = "camelot_v3"
	CurvePool        PoolType = "curve"
	BalancerV2Pool   PoolType = "balancer_v2"
	BalancerV3Pool   PoolType = "balancer_v3"
	RamsesPool       PoolType = "ramses"
	KyberClassicPool PoolType = "kyber_classic"
	KyberElasticPool PoolType = "kyber_elastic"
	SushiSwapPool    PoolType = "sushiswap"
	TraderJoePool    PoolType = "trader_joe"
)

// Transaction represents a decoded transaction from the sequencer
type Transaction struct {
	Hash        Hash            `json:"hash"`
	BlockNumber uint64          `json:"block_number"`
	Timestamp   time.Time       `json:"timestamp"`
	From        Address         `json:"from"`
	To          Address         `json:"to"`
	Value       *big.Int        `json:"value"`
	Data        []byte          `json:"data"`
	GasLimit    uint64          `json:"gas_limit"`
	GasPrice    *big.Int        `json:"gas_price"`
	Actions     []DecodedAction `json:"actions"` // Decoded DEX actions
}

// DecodedAction represents a decoded DEX action from a transaction
type DecodedAction struct {
	Type      ActionType             `json:"type"`
	Protocol  ProtocolType           `json:"protocol"`
	Pool      Address                `json:"pool"`
	TokenIn   Token                  `json:"token_in"`
	TokenOut  Token                  `json:"token_out"`
	AmountIn  *big.Int               `json:"amount_in"`
	AmountOut *big.Int               `json:"amount_out"`
	Path      []Token                `json:"path,omitempty"`   // For multi-hop swaps
	Params    map[string]interface{} `json:"params,omitempty"` // Protocol-specific parameters
	Error     error                  `json:"error,omitempty"`  // Decoding error if any
}

// ActionType represents the type of DEX action
type ActionType string

const (
	SwapAction            ActionType = "swap"
	AddLiquidityAction    ActionType = "add_liquidity"
	RemoveLiquidityAction ActionType = "remove_liquidity"
	FlashLoanAction       ActionType = "flash_loan"
)

// ProtocolType represents the DEX protocol
type ProtocolType string

const (
	UniswapV2Protocol    ProtocolType = "uniswap_v2"
	UniswapV3Protocol    ProtocolType = "uniswap_v3"
	CamelotV3Protocol    ProtocolType = "camelot_v3"
	CurveProtocol        ProtocolType = "curve"
	BalancerV2Protocol   ProtocolType = "balancer_v2"
	BalancerV3Protocol   ProtocolType = "balancer_v3"
	RamsesProtocol       ProtocolType = "ramses"
	KyberClassicProtocol ProtocolType = "kyber_classic"
	KyberElasticProtocol ProtocolType = "kyber_elastic"
	// Aggregators
	OneInchProtocol   ProtocolType = "1inch"
	OpenOceanProtocol ProtocolType = "openocean"
	OdosProtocol      ProtocolType = "odos"
	ParaswapProtocol  ProtocolType = "paraswap"
	ZeroXProtocol     ProtocolType = "0x"
	// Additional DEXes
	SushiSwapProtocol ProtocolType = "sushiswap"
	GMXProtocol       ProtocolType = "gmx"
	TraderJoeProtocol ProtocolType = "trader_joe"
	WooFiProtocol     ProtocolType = "woofi"
)

// Config holds the application configuration
type Config struct {
	RPCEndpoint    string   `json:"rpc_endpoint"`
	SequencerURL   string   `json:"sequencer_url"`
	LogLevel       string   `json:"log_level"`
	MaxConcurrency int      `json:"max_concurrency"`
	DBPath         string   `json:"db_path"`
	MaxGasPrice    *big.Int `json:"max_gas_price"`
	MinProfit      *big.Int `json:"min_profit"` // Minimum profit in USD (scaled by 1e18)
	MevRelayURL    string   `json:"mev_relay_url"`
}

// SecureConfig holds sensitive configuration that should never be serialized
// These values should be loaded from environment variables or a secure vault
type SecureConfig struct {
	// PrivateKey for transaction signing - NEVER log or serialize this
	PrivateKey string
	// FlashbotsAuthKey for MEV relay authentication
	FlashbotsAuthKey string
}

// RedactedString returns a redacted version for logging
func (s *SecureConfig) RedactedString() string {
	pkStatus := "not set"
	if s.PrivateKey != "" {
		pkStatus = "set (redacted)"
	}
	fbStatus := "not set"
	if s.FlashbotsAuthKey != "" {
		fbStatus = "set (redacted)"
	}
	return "SecureConfig{PrivateKey: " + pkStatus + ", FlashbotsAuthKey: " + fbStatus + "}"
}

// MarshalJSON prevents accidental serialization of sensitive data
func (s *SecureConfig) MarshalJSON() ([]byte, error) {
	return []byte(`{"error":"SecureConfig cannot be serialized"}`), nil
}

// SwapSimulation represents the result of simulating a swap
type SwapSimulation struct {
	TokenIn           Token    `json:"token_in"`
	TokenOut          Token    `json:"token_out"`
	AmountIn          *big.Int `json:"amount_in"`
	ExpectedAmountOut *big.Int `json:"expected_amount_out"`
	GasEstimate       uint64   `json:"gas_estimate"`
	PriceImpact       float64  `json:"price_impact"` // Percentage
	Error             error    `json:"error,omitempty"`
}

// ArbitrageOpportunity represents a detected arbitrage opportunity
type ArbitrageOpportunity struct {
	ID            string             `json:"id"`
	Type          ArbitrageType      `json:"type"`
	Profit        *big.Int           `json:"profit"` // Expected profit in USD (scaled by 1e18)
	ProfitToken   Token              `json:"profit_token"`
	ExecutionPath []DecodedAction    `json:"execution_path"`
	EstimatedGas  uint64             `json:"estimated_gas"`
	Probability   float64            `json:"probability"` // 0.0 to 1.0
	RiskFactor    float64            `json:"risk_factor"` // 0.0 to 1.0
	Expiration    time.Time          `json:"expiration"`
	SourceTx      *types.Transaction `json:"source_tx,omitempty"` // Original transaction that triggered this opportunity
}

// ArbitrageType represents the type of arbitrage opportunity
type ArbitrageType string

const (
	TriangleArbitrageType    ArbitrageType = "triangle"
	CrossDEXArbitrageType    ArbitrageType = "cross_dex"
	FlashArbitrageType       ArbitrageType = "flash"
	StatisticalArbitrageType ArbitrageType = "statistical"
)

// PoolState represents the current state of a pool
type PoolState struct {
	Reserves0       *big.Int          `json:"reserves_0"`
	Reserves1       *big.Int          `json:"reserves_1"`
	Reserves2       *big.Int          `json:"reserves_2,omitempty"` // For 3-token pools
	Liquidity       *big.Int          `json:"liquidity"`
	Price0To1       float64           `json:"price_0_to_1"` // Price of token0 in terms of token1
	Price1To0       float64           `json:"price_1_to_0"` // Price of token1 in terms of token0
	LastUpdated     time.Time         `json:"last_updated"`
	Tick            int               `json:"tick,omitempty"`             // For Uniswap V3 style pools
	ActiveLiquidity *big.Int          `json:"active_liquidity,omitempty"` // For concentrated liquidity pools
	ExtraData       map[string]string `json:"extra_data,omitempty"`       // Protocol-specific data
}
