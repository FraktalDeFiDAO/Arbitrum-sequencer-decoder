// Package blockchain provides utilities for querying blockchain data
package blockchain

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"net/url"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Default configuration values
const (
	DefaultRequestsPerSecond = 10
	DefaultTimeout           = 30 * time.Second
	DefaultMaxRetries        = 3
)

// ClientConfig holds configuration for the blockchain client
type ClientConfig struct {
	RPCURL            string        // RPC endpoint URL
	RequestsPerSecond int           // Rate limit (requests per second)
	Timeout           time.Duration // Default timeout for operations
	MaxRetries        int           // Maximum retries on transient errors
}

// DefaultClientConfig returns sensible defaults for client configuration
func DefaultClientConfig(rpcURL string) *ClientConfig {
	return &ClientConfig{
		RPCURL:            rpcURL,
		RequestsPerSecond: DefaultRequestsPerSecond,
		Timeout:           DefaultTimeout,
		MaxRetries:        DefaultMaxRetries,
	}
}

// ValidateConfig validates the client configuration
func (c *ClientConfig) ValidateConfig() error {
	if c.RPCURL == "" {
		return fmt.Errorf("RPC URL cannot be empty")
	}

	if _, err := url.Parse(c.RPCURL); err != nil {
		return fmt.Errorf("invalid RPC URL: %w", err)
	}

	if c.RequestsPerSecond <= 0 {
		c.RequestsPerSecond = DefaultRequestsPerSecond
	}

	if c.Timeout <= 0 {
		c.Timeout = DefaultTimeout
	}

	if c.MaxRetries < 0 {
		c.MaxRetries = DefaultMaxRetries
	}

	return nil
}

// RateLimiter provides token-bucket based rate limiting
type RateLimiter struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter with the given requests per second
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	return &RateLimiter{
		tokens:     float64(requestsPerSecond),
		maxTokens:  float64(requestsPerSecond),
		refillRate: float64(requestsPerSecond),
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available or context is cancelled
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		r.mu.Lock()
		// Refill tokens based on elapsed time
		now := time.Now()
		elapsed := now.Sub(r.lastRefill).Seconds()
		r.tokens += elapsed * r.refillRate
		if r.tokens > r.maxTokens {
			r.tokens = r.maxTokens
		}
		r.lastRefill = now

		if r.tokens >= 1.0 {
			r.tokens -= 1.0
			r.mu.Unlock()
			return nil
		}

		// Calculate wait time for next token
		waitTime := time.Duration((1.0 - r.tokens) / r.refillRate * float64(time.Second))
		r.mu.Unlock()

		// Use a timer instead of time.After for proper cleanup
		timer := time.NewTimer(waitTime)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			// Continue loop to try again
		}
	}
}

// Client is a wrapper around ethclient.Client with rate limiting and additional utility methods
type Client struct {
	*ethclient.Client
	rateLimiter *RateLimiter
	config      *ClientConfig
}

// NewClient creates a new blockchain client with default configuration
func NewClient(rpcURL string) (*Client, error) {
	return NewClientWithConfig(DefaultClientConfig(rpcURL))
}

// NewClientWithConfig creates a new blockchain client with custom configuration
func NewClientWithConfig(config *ClientConfig) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if err := config.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := ethclient.Dial(config.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	return &Client{
		Client:      client,
		rateLimiter: NewRateLimiter(config.RequestsPerSecond),
		config:      config,
	}, nil
}

// withRateLimit wraps an operation with rate limiting
func (c *Client) withRateLimit(ctx context.Context) error {
	if c.rateLimiter == nil {
		return nil
	}
	return c.rateLimiter.Wait(ctx)
}

// withTimeout ensures context has a timeout, adding one if not present
func (c *Client) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, func() {} // no-op cancel since deadline already set
	}
	return context.WithTimeout(ctx, c.config.Timeout)
}

// withRetry wraps an operation with retry logic using exponential backoff
func (c *Client) withRetry(ctx context.Context, operation func() error) error {
	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if err := operation(); err != nil {
			lastErr = err
			// Don't retry on context cancellation
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// Don't retry on the last attempt
			if attempt == c.config.MaxRetries {
				break
			}
			// Exponential backoff: 100ms, 200ms, 400ms, ...
			backoff := time.Duration(100<<attempt) * time.Millisecond
			if backoff > 5*time.Second {
				backoff = 5 * time.Second
			}
			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
				// Continue to next attempt
			}
		} else {
			return nil
		}
	}
	return fmt.Errorf("operation failed after %d retries: %w", c.config.MaxRetries+1, lastErr)
}

// BlockNumber returns the current block number with rate limiting and retry
func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	var result uint64
	err := c.withRetry(ctx, func() error {
		if err := c.withRateLimit(ctx); err != nil {
			return fmt.Errorf("rate limit wait failed: %w", err)
		}
		reqCtx, cancel := c.withTimeout(ctx)
		defer cancel()
		var err error
		result, err = c.Client.BlockNumber(reqCtx)
		return err
	})
	return result, err
}

// HeaderByNumber returns the header for the given block number with rate limiting and retry
func (c *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	var result *types.Header
	err := c.withRetry(ctx, func() error {
		if err := c.withRateLimit(ctx); err != nil {
			return fmt.Errorf("rate limit wait failed: %w", err)
		}
		reqCtx, cancel := c.withTimeout(ctx)
		defer cancel()
		var err error
		result, err = c.Client.HeaderByNumber(reqCtx, number)
		return err
	})
	return result, err
}

// BlockByNumber returns the block for the given number with rate limiting and retry
func (c *Client) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	var result *types.Block
	err := c.withRetry(ctx, func() error {
		if err := c.withRateLimit(ctx); err != nil {
			return fmt.Errorf("rate limit wait failed: %w", err)
		}
		reqCtx, cancel := c.withTimeout(ctx)
		defer cancel()
		var err error
		result, err = c.Client.BlockByNumber(reqCtx, number)
		return err
	})
	return result, err
}

// BlockByHash returns the block for the given hash with rate limiting and retry
func (c *Client) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	var result *types.Block
	err := c.withRetry(ctx, func() error {
		if err := c.withRateLimit(ctx); err != nil {
			return fmt.Errorf("rate limit wait failed: %w", err)
		}
		reqCtx, cancel := c.withTimeout(ctx)
		defer cancel()
		var err error
		result, err = c.Client.BlockByHash(reqCtx, hash)
		return err
	})
	return result, err
}

// TransactionByHash returns the transaction for the given hash with rate limiting and retry
func (c *Client) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	var result *types.Transaction
	var isPending bool
	err := c.withRetry(ctx, func() error {
		if err := c.withRateLimit(ctx); err != nil {
			return fmt.Errorf("rate limit wait failed: %w", err)
		}
		reqCtx, cancel := c.withTimeout(ctx)
		defer cancel()
		var err error
		result, isPending, err = c.Client.TransactionByHash(reqCtx, hash)
		return err
	})
	return result, isPending, err
}

// TransactionReceipt returns the receipt for the given transaction hash with rate limiting and retry
func (c *Client) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	var result *types.Receipt
	err := c.withRetry(ctx, func() error {
		if err := c.withRateLimit(ctx); err != nil {
			return fmt.Errorf("rate limit wait failed: %w", err)
		}
		reqCtx, cancel := c.withTimeout(ctx)
		defer cancel()
		var err error
		result, err = c.Client.TransactionReceipt(reqCtx, txHash)
		return err
	})
	return result, err
}

// FilterLogs filters logs with rate limiting and retry
func (c *Client) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	var result []types.Log
	err := c.withRetry(ctx, func() error {
		if err := c.withRateLimit(ctx); err != nil {
			return fmt.Errorf("rate limit wait failed: %w", err)
		}
		reqCtx, cancel := c.withTimeout(ctx)
		defer cancel()
		var err error
		result, err = c.Client.FilterLogs(reqCtx, query)
		return err
	})
	return result, err
}

// CodeAt returns the code at the given address with rate limiting and retry
func (c *Client) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	var result []byte
	err := c.withRetry(ctx, func() error {
		if err := c.withRateLimit(ctx); err != nil {
			return fmt.Errorf("rate limit wait failed: %w", err)
		}
		reqCtx, cancel := c.withTimeout(ctx)
		defer cancel()
		var err error
		result, err = c.Client.CodeAt(reqCtx, account, blockNumber)
		return err
	})
	return result, err
}

// Transaction represents a blockchain transaction with additional metadata
type Transaction struct {
	Hash        common.Hash     `json:"hash"`
	BlockHash   common.Hash     `json:"block_hash,omitempty"`
	BlockNumber *big.Int        `json:"block_number"`
	Index       uint            `json:"index"`
	From        common.Address  `json:"from"`
	To          *common.Address `json:"to"`
	Value       *big.Int        `json:"value"`
	Input       []byte          `json:"input"`
	Gas         uint64          `json:"gas"`
	GasPrice    *big.Int        `json:"gas_price"`
	Nonce       uint64          `json:"nonce"`
	Type        uint8           `json:"type"`
	Receipt     *types.Receipt  `json:"receipt,omitempty"`
}

// ContractFilter defines criteria for filtering contract transactions
type ContractFilter struct {
	Addresses  []common.Address
	Signatures []common.Hash
	FromBlock  *big.Int
	ToBlock    *big.Int
}

// GetTransactionsByContract retrieves transactions to specific contract addresses
func (c *Client) GetTransactionsByContract(ctx context.Context, filter ContractFilter, limit int) ([]Transaction, error) {
	var allTxs []Transaction

	// If no block range is specified, try to get recent blocks
	if filter.FromBlock == nil {
		header, err := c.HeaderByNumber(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest block: %w", err)
		}
		// Look at the last 1000 blocks by default
		filter.FromBlock = new(big.Int).Sub(header.Number, big.NewInt(1000))
		if filter.FromBlock.Sign() < 0 {
			filter.FromBlock = big.NewInt(0)
		}
	}

	if filter.ToBlock == nil {
		header, err := c.HeaderByNumber(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest block: %w", err)
		}
		filter.ToBlock = header.Number
	}

	// Query logs first to identify transactions of interest
	query := ethereum.FilterQuery{
		FromBlock: filter.FromBlock,
		ToBlock:   filter.ToBlock,
		Addresses: filter.Addresses,
		Topics:    [][]common.Hash{},
	}

	// If specific function signatures are provided, add them to the filter
	if len(filter.Signatures) > 0 {
		query.Topics = append(query.Topics, filter.Signatures)
	}

	logs, err := c.FilterLogs(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %w", err)
	}

	// Extract unique transaction hashes
	txHashes := make(map[common.Hash]bool)
	for _, logEntry := range logs {
		txHashes[logEntry.TxHash] = true
	}

	// Fetch transactions by hash
	count := 0
	for txHash := range txHashes {
		if limit > 0 && count >= limit {
			break
		}

		tx, isPending, err := c.TransactionByHash(ctx, txHash)
		if err != nil {
			log.Printf("Could not get transaction %s: %v", txHash.Hex(), err)
			continue
		}

		if isPending {
			log.Printf("Skipping pending transaction %s", txHash.Hex())
			continue
		}

		// Get receipt
		receipt, err := c.TransactionReceipt(ctx, txHash)
		if err != nil {
			log.Printf("Could not get receipt for transaction %s: %v", txHash.Hex(), err)
		}

		// Get sender
		from, err := types.Sender(types.NewLondonSigner(tx.ChainId()), tx)
		if err != nil {
			log.Printf("Could not get sender for transaction %s: %v", txHash.Hex(), err)
		}

		// Get block details
		block, err := c.BlockByHash(ctx, receipt.BlockHash)
		if err != nil || block == nil {
			log.Printf("Could not get block for transaction %s: %v", txHash.Hex(), err)
			// Skip this transaction if we can't get the block
			continue
		}

		transaction := Transaction{
			Hash:        tx.Hash(),
			BlockHash:   receipt.BlockHash,
			BlockNumber: block.Number(),
			Index:       receipt.TransactionIndex,
			From:        from,
			To:          tx.To(),
			Value:       tx.Value(),
			Input:       tx.Data(),
			Gas:         tx.Gas(),
			GasPrice:    tx.GasPrice(),
			Nonce:       tx.Nonce(),
			Type:        uint8(tx.Type()),
			Receipt:     receipt,
		}

		allTxs = append(allTxs, transaction)
		count++
	}

	return allTxs, nil
}

// GetRecentTransactionsByAddress gets recent transactions to a specific address
func (c *Client) GetRecentTransactionsByAddress(ctx context.Context, address common.Address, limit int) ([]Transaction, error) {
	currentBlock, err := c.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current block: %w", err)
	}

	// Look at the last 10000 blocks for recent transactions
	fromBlock := new(big.Int).Sub(new(big.Int).SetUint64(currentBlock), big.NewInt(10000))
	if fromBlock.Sign() < 0 {
		fromBlock = big.NewInt(0)
	}

	filter := ContractFilter{
		Addresses: []common.Address{address},
		FromBlock: fromBlock,
		ToBlock:   new(big.Int).SetUint64(currentBlock),
	}

	return c.GetTransactionsByContract(ctx, filter, limit)
}

// IsValidAddress checks if an address is valid and potentially a smart contract
func (c *Client) IsValidAddress(ctx context.Context, address common.Address) (bool, error) {
	// Try to get the bytecode at the address
	_, err := c.CodeAt(ctx, address, nil)
	if err != nil {
		return false, err
	}

	// If bytecode is not empty, it's likely a contract
	// If it's empty, it could be either an EOA or a contract that returned empty bytecode
	return true, nil
}

// GetContractABI attempts to retrieve the ABI of a contract (this would require external service in practice)
func GetContractABI(contractAddress string) (*abi.ABI, error) {
	// This is a simplified function
	// In practice, you might query a service like Etherscan API for ABI
	// For this implementation, we'll return an error indicating it's not implemented
	return nil, fmt.Errorf("getting contract ABI from %s not implemented - consider using an external service", contractAddress)
}

// WaitUntilBlock waits until a specific block number is reached
func (c *Client) WaitUntilBlock(ctx context.Context, targetBlock *big.Int) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			currentBlock, err := c.BlockNumber(ctx)
			if err != nil {
				return fmt.Errorf("error getting current block: %w", err)
			}

			if currentBlock >= targetBlock.Uint64() {
				return nil
			}
		}
	}
}

// GetBlockTransactions retrieves all transactions from a specific block
func (c *Client) GetBlockTransactions(ctx context.Context, blockNumber *big.Int) ([]Transaction, error) {
	block, err := c.BlockByNumber(ctx, blockNumber)
	if err != nil || block == nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	var transactions []Transaction

	for i, tx := range block.Transactions() {
		receipt, err := c.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			log.Printf("Could not get receipt for transaction %d in block %d: %v", i, blockNumber, err)
		}

		// Get sender
		from, err := types.Sender(types.NewLondonSigner(tx.ChainId()), tx)
		if err != nil {
			log.Printf("Could not get sender for transaction %d in block %d: %v", i, blockNumber, err)
		}

		transaction := Transaction{
			Hash:        tx.Hash(),
			BlockHash:   block.Hash(),
			BlockNumber: block.Number(),
			Index:       uint(i),
			From:        from,
			To:          tx.To(),
			Value:       tx.Value(),
			Input:       tx.Data(),
			Gas:         tx.Gas(),
			GasPrice:    tx.GasPrice(),
			Nonce:       tx.Nonce(),
			Type:        uint8(tx.Type()),
			Receipt:     receipt,
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// GetTransactionDetails gets full details for a specific transaction
func (c *Client) GetTransactionDetails(ctx context.Context, txHash common.Hash) (*Transaction, error) {
	tx, isPending, err := c.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if isPending {
		return nil, fmt.Errorf("transaction %s is still pending", txHash.Hex())
	}

	receipt, err := c.TransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	// Get block details
	block, err := c.BlockByHash(ctx, receipt.BlockHash)
	if err != nil || block == nil {
		return nil, fmt.Errorf("failed to get block for transaction: %w", err)
	}

	// Get sender
	from, err := types.Sender(types.NewLondonSigner(tx.ChainId()), tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender: %w", err)
	}

	transaction := &Transaction{
		Hash:        tx.Hash(),
		BlockHash:   receipt.BlockHash,
		BlockNumber: block.Number(),
		Index:       receipt.TransactionIndex,
		From:        from,
		To:          tx.To(),
		Value:       tx.Value(),
		Input:       tx.Data(),
		Gas:         tx.Gas(),
		GasPrice:    tx.GasPrice(),
		Nonce:       tx.Nonce(),
		Type:        uint8(tx.Type()),
		Receipt:     receipt,
	}

	return transaction, nil
}
