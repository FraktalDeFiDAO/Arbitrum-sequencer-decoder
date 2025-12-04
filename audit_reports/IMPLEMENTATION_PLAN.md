# Implementation Plan for Audit Fixes

**Project:** Arbitrum Sequencer Decoder
**Date:** 2025-12-02
**Based on:** COMPREHENSIVE_AUDIT_REPORT.md

---

## Overview

This document provides detailed implementation guidance for fixing issues identified in the comprehensive audit. Each fix includes the specific code changes required, testing strategy, and verification steps.

---

## Phase 1: Critical Fixes (Priority P0)

### Task 1.1: Fix ABI Parameter Unpacking in Uniswap V3 Decoder

**Issue Reference:** Audit Report 1.1
**Estimated Time:** 4 hours
**File:** `pkg/decoder/uniswap_v3/decoder.go`

#### Implementation Steps

1. **Update `decodeExactInput` function:**

```go
// decodeExactInput decodes the exactInput function parameters
func (d *UniswapV3Decoder) decodeExactInput(calldata []byte) ([]types.DecodedAction, error) {
    args := RouterABIInstance.Methods["exactInput"]

    v, err := args.Inputs.UnpackValues(calldata[4:])
    if err != nil {
        return nil, fmt.Errorf("failed to unpack exactInput parameters: %w", err)
    }

    if len(v) == 0 {
        return nil, errors.New("no parameters unpacked from exactInput")
    }

    // The ABI unpacks a tuple - need to access it properly
    // Use reflect or type assertion based on go-ethereum ABI behavior
    paramsMap, ok := v[0].(map[string]interface{})
    if !ok {
        // Alternative: try struct approach
        return d.decodeExactInputFromStruct(v[0])
    }

    pathBytes, ok := paramsMap["path"].([]byte)
    if !ok {
        return nil, errors.New("invalid path parameter in exactInput")
    }

    path, err := DecodePath(pathBytes)
    if err != nil {
        return nil, fmt.Errorf("failed to decode path: %w", err)
    }

    if len(path) < 2 {
        return nil, errors.New("path must have at least 2 tokens")
    }

    recipient, _ := paramsMap["recipient"].(common.Address)
    amountIn, _ := paramsMap["amountIn"].(*big.Int)
    amountOutMinimum, _ := paramsMap["amountOutMinimum"].(*big.Int)
    deadline, _ := paramsMap["deadline"].(*big.Int)

    var actions []types.DecodedAction
    for i := 0; i < len(path)-1; i++ {
        action := types.DecodedAction{
            Type:      types.SwapAction,
            Protocol:  types.UniswapV3Protocol,
            TokenIn:   types.Token{Address: path[i]},
            TokenOut:  types.Token{Address: path[i+1]},
            AmountIn:  amountIn,
            AmountOut: amountOutMinimum,
            Params: map[string]interface{}{
                "function":         "exactInput",
                "recipient":        recipient,
                "deadline":         deadline,
                "amountOutMinimum": amountOutMinimum,
                "pathIndex":        i,
                "totalHops":        len(path) - 1,
            },
        }
        actions = append(actions, action)
    }

    return actions, nil
}

// Helper for struct-based unpacking
func (d *UniswapV3Decoder) decodeExactInputFromStruct(v interface{}) ([]types.DecodedAction, error) {
    // Use reflection to handle the struct
    val := reflect.ValueOf(v)
    if val.Kind() == reflect.Ptr {
        val = val.Elem()
    }

    pathField := val.FieldByName("Path")
    if !pathField.IsValid() {
        return nil, errors.New("Path field not found in struct")
    }

    pathBytes, ok := pathField.Interface().([]byte)
    if !ok {
        return nil, errors.New("invalid Path type")
    }

    // Continue with path decoding...
    path, err := DecodePath(pathBytes)
    if err != nil {
        return nil, err
    }

    // Build actions similar to above
    // ...
    return nil, errors.New("not fully implemented")
}
```

2. **Apply similar fixes to other decode functions:**
   - `decodeExactOutput`
   - `decodeExactInputSingle`
   - `decodeExactOutputSingle`

#### Testing Strategy

```go
func TestDecodeExactInput_RealData(t *testing.T) {
    decoder := NewUniswapV3Decoder()

    // Use real calldata from Arbitrum mainnet
    realCalldata := common.FromHex("c04b8d59000000000000...")

    routerAddr := common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564")
    tx := testhelpers.CreateTestTransaction(routerAddr, realCalldata, nil)

    actions, err := decoder.Decode(tx, routerAddr.Hex())
    require.NoError(t, err)
    require.NotEmpty(t, actions)

    // Verify decoded values match expected
    assert.Equal(t, types.SwapAction, actions[0].Type)
    assert.NotEqual(t, common.Address{}, actions[0].TokenIn.Address)
}
```

---

### Task 1.2: Correct Function Signatures

**Issue Reference:** Audit Report 1.2
**Estimated Time:** 2 hours
**File:** `pkg/decoder/uniswap_v3/abi_bindings.go`

#### Implementation Steps

1. **Update the FunctionSignatures slice:**

```go
// Function signatures for Uniswap V3 swap operations
// Verified against Uniswap V3 SwapRouter contract on Arbitrum
var FunctionSignatures = []string{
    "0xc04b8d59", // exactInput(ExactInputParams)
    "0xf28c0498", // exactOutput(ExactOutputParams)
    "0x414bf389", // exactInputSingle(ExactInputSingleParams)
    "0xdb3e2198", // exactOutputSingle(ExactOutputSingleParams)
    "0xac9650d8", // multicall(bytes[])
    "0x5ae401dc", // multicall(uint256,bytes[])
    "0x1f0464d1", // multicall(bytes32,bytes[])
}
```

2. **Update the switch statement in Decode:**

```go
func (d *UniswapV3Decoder) Decode(tx *ethtypes.Transaction, toAddress string) ([]types.DecodedAction, error) {
    calldata := tx.Data()
    if len(calldata) < 4 {
        return nil, errors.New("calldata too short")
    }

    signature := "0x" + common.Bytes2Hex(calldata[:4])

    switch signature {
    case "0xc04b8d59": // exactInput
        return d.decodeExactInput(calldata)
    case "0xf28c0498": // exactOutput
        return d.decodeExactOutput(calldata)
    case "0x414bf389": // exactInputSingle
        return d.decodeExactInputSingle(calldata)
    case "0xdb3e2198": // exactOutputSingle
        return d.decodeExactOutputSingle(calldata)
    case "0xac9650d8", "0x5ae401dc", "0x1f0464d1": // multicall variants
        return d.decodeMulticall(calldata)
    default:
        return nil, fmt.Errorf("unknown Uniswap V3 function signature: %s", signature)
    }
}
```

3. **Add multicall decoder:**

```go
// decodeMulticall decodes multicall batched transactions
func (d *UniswapV3Decoder) decodeMulticall(calldata []byte) ([]types.DecodedAction, error) {
    // Skip the 4-byte selector
    data := calldata[4:]

    // Multicall format: bytes[] array of encoded calls
    // Each element is an encoded function call that we can recursively decode

    var allActions []types.DecodedAction

    // Parse the ABI-encoded bytes array
    // This requires proper ABI decoding of the dynamic array

    // Simplified implementation - actual implementation needs proper ABI parsing
    return allActions, nil
}
```

#### Verification

```bash
# Verify signatures match actual contract
cast sig "exactInput((bytes,address,uint256,uint256,uint256))"
# Should output: 0xc04b8d59

cast sig "exactInputSingle((address,address,uint24,address,uint256,uint256,uint256,uint160))"
# Should output: 0x414bf389
```

---

### Task 1.3: Fix Integer Overflow in Liquidity Tracking

**Issue Reference:** Audit Report 1.3
**Estimated Time:** 1 hour
**File:** `agents/auditing/decoder_auditor.go`

#### Implementation Steps

1. **Change TotalLiquidity to use big.Int:**

```go
// DecoderStats holds statistics about a decoder's performance
type DecoderStats struct {
    TotalDecoded    int64
    TotalFailed     int64
    AvgDecodeTime   time.Duration
    LastDecodeTime  time.Duration
    SuccessRate     float64
    LastUpdated     time.Time
    TotalLiquidity  *big.Int          // Changed from int64
    MaxDecodeTime   time.Duration
    MinDecodeTime   time.Duration
    ErrorDetails    []string
}
```

2. **Update the updateStats function:**

```go
func (da *DecoderAuditor) updateStats(decoderName string, success bool, decodeTime time.Duration, actions []types.DecodedAction) {
    da.mu.Lock()
    defer da.mu.Unlock()

    stats, exists := da.stats[decoderName]
    if !exists {
        stats = &DecoderStats{
            TotalDecoded:   0,
            TotalFailed:    0,
            AvgDecodeTime:  0,
            MaxDecodeTime:  0,
            MinDecodeTime:  time.Duration(1<<63 - 1),
            SuccessRate:    0,
            TotalLiquidity: big.NewInt(0),  // Initialize as big.Int
            ErrorDetails:   make([]string, 0),
        }
        da.stats[decoderName] = stats
    }

    // ... existing success/failure logic ...

    // Safe liquidity accumulation
    for _, action := range actions {
        if action.AmountIn != nil && action.AmountIn.Sign() > 0 {
            stats.TotalLiquidity = new(big.Int).Add(stats.TotalLiquidity, action.AmountIn)
        }
        if action.AmountOut != nil && action.AmountOut.Sign() > 0 {
            stats.TotalLiquidity = new(big.Int).Add(stats.TotalLiquidity, action.AmountOut)
        }
    }

    stats.LastUpdated = time.Now()
}
```

3. **Update GetStats to return a copy:**

```go
func (da *DecoderAuditor) GetStats() map[string]*DecoderStats {
    da.mu.RLock()
    defer da.mu.RUnlock()

    result := make(map[string]*DecoderStats)
    for name, stats := range da.stats {
        errorDetails := make([]string, len(stats.ErrorDetails))
        copy(errorDetails, stats.ErrorDetails)

        result[name] = &DecoderStats{
            TotalDecoded:   stats.TotalDecoded,
            TotalFailed:    stats.TotalFailed,
            AvgDecodeTime:  stats.AvgDecodeTime,
            MaxDecodeTime:  stats.MaxDecodeTime,
            MinDecodeTime:  stats.MinDecodeTime,
            LastDecodeTime: stats.LastDecodeTime,
            SuccessRate:    stats.SuccessRate,
            LastUpdated:    stats.LastUpdated,
            TotalLiquidity: new(big.Int).Set(stats.TotalLiquidity),  // Copy
            ErrorDetails:   errorDetails,
        }
    }
    return result
}
```

---

## Phase 2: High Priority Fixes (Priority P1)

### Task 2.1: Safe ABI Initialization

**Issue Reference:** Audit Report 2.1
**Estimated Time:** 3 hours
**Files:** `pkg/decoder/uniswap_v3/abi_bindings.go`, `pkg/decoder/camelot_v3/abi_bindings.go`

#### Implementation

```go
package uniswap_v3

import (
    "sync"
    "strings"
    "github.com/ethereum/go-ethereum/accounts/abi"
)

var (
    routerABIOnce     sync.Once
    routerABI         abi.ABI
    routerABIInitErr  error
)

// GetRouterABI returns the parsed Uniswap V3 Router ABI
// Thread-safe and returns any initialization error
func GetRouterABI() (abi.ABI, error) {
    routerABIOnce.Do(func() {
        routerABI, routerABIInitErr = abi.JSON(strings.NewReader(RouterABI))
    })
    return routerABI, routerABIInitErr
}

// MustGetRouterABI returns the Router ABI or panics
// Only use during application startup when ABI errors should be fatal
func MustGetRouterABI() abi.ABI {
    abi, err := GetRouterABI()
    if err != nil {
        panic(fmt.Sprintf("failed to initialize Uniswap V3 Router ABI: %v", err))
    }
    return abi
}
```

Update decoder to use the safe getter:

```go
func (d *UniswapV3Decoder) decodeExactInput(calldata []byte) ([]types.DecodedAction, error) {
    routerABI, err := GetRouterABI()
    if err != nil {
        return nil, fmt.Errorf("ABI not initialized: %w", err)
    }

    args := routerABI.Methods["exactInput"]
    // ... rest of implementation
}
```

---

### Task 2.2: Fix Recursive Price Impact Calculation

**Issue Reference:** Audit Report 2.2
**Estimated Time:** 2 hours
**File:** `pkg/simulator/camelot_v3/math.go`

#### Implementation

```go
// SimulateSwap simulates a swap in a Camelot V3 pool
// Set calculatePriceImpact to false to prevent recursion
func SimulateSwap(poolState *PoolState, tokenIn common.Address, amountIn *big.Int, calculateImpact bool) (*types.SwapSimulation, error) {
    if poolState.Reserves0.Sign() <= 0 || poolState.Reserves1.Sign() <= 0 {
        return nil, errors.New("invalid pool reserves: must be positive")
    }

    if amountIn.Sign() <= 0 {
        return nil, errors.New("amount in must be positive")
    }

    var expectedAmountOut *big.Int

    if tokenIn == poolState.Token0 {
        expectedAmountOut = calculateAmountOut(amountIn, poolState.Reserves0, poolState.Reserves1)
    } else if tokenIn == poolState.Token1 {
        expectedAmountOut = calculateAmountOut(amountIn, poolState.Reserves1, poolState.Reserves0)
    } else {
        return nil, errors.New("token not in pool")
    }

    var priceImpact float64
    if calculateImpact {
        priceImpact = calculatePriceImpactDirect(poolState, tokenIn, amountIn, expectedAmountOut)
    }

    simulation := &types.SwapSimulation{
        TokenIn:           types.Token{Address: tokenIn},
        TokenOut:          types.Token{Address: getOtherToken(poolState, tokenIn)},
        AmountIn:          amountIn,
        ExpectedAmountOut: expectedAmountOut,
        GasEstimate:       100000,
        PriceImpact:       priceImpact,
    }

    return simulation, nil
}

// calculatePriceImpactDirect calculates price impact without calling SimulateSwap
func calculatePriceImpactDirect(poolState *PoolState, tokenIn common.Address, amountIn, amountOut *big.Int) float64 {
    // Calculate current price
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

    // Calculate new reserves after swap
    var newReservesIn, newReservesOut *big.Int
    if tokenIn == poolState.Token0 {
        newReservesIn = new(big.Int).Add(poolState.Reserves0, amountIn)
        newReservesOut = new(big.Int).Sub(poolState.Reserves1, amountOut)
    } else {
        newReservesIn = new(big.Int).Add(poolState.Reserves1, amountIn)
        newReservesOut = new(big.Int).Sub(poolState.Reserves0, amountOut)
    }

    // Calculate new price
    newPrice := new(big.Float).Quo(
        new(big.Float).SetInt(newReservesOut),
        new(big.Float).SetInt(newReservesIn),
    )

    // Calculate impact as percentage
    priceDiff := new(big.Float).Sub(currentPrice, newPrice)
    priceImpact := new(big.Float).Quo(priceDiff, currentPrice)
    result, _ := priceImpact.Float64()

    return result * 100
}
```

---

### Task 2.3: Add RPC Rate Limiting

**Issue Reference:** Audit Report 2.3
**Estimated Time:** 4 hours
**File:** `pkg/blockchain/client.go`

#### Implementation

```go
package blockchain

import (
    "context"
    "time"

    "golang.org/x/time/rate"
    "github.com/ethereum/go-ethereum/ethclient"
)

// ClientConfig holds configuration for the blockchain client
type ClientConfig struct {
    RPCURL         string
    RequestsPerSec int           // Rate limit (requests per second)
    Timeout        time.Duration // Default timeout for operations
    MaxRetries     int           // Maximum retries on transient errors
}

// DefaultClientConfig returns sensible defaults
func DefaultClientConfig(rpcURL string) *ClientConfig {
    return &ClientConfig{
        RPCURL:         rpcURL,
        RequestsPerSec: 10,
        Timeout:        30 * time.Second,
        MaxRetries:     3,
    }
}

// Client is a wrapper around ethclient.Client with rate limiting
type Client struct {
    *ethclient.Client
    rateLimiter *rate.Limiter
    config      *ClientConfig
}

// NewClient creates a new rate-limited blockchain client
func NewClient(config *ClientConfig) (*Client, error) {
    if config == nil {
        return nil, errors.New("config cannot be nil")
    }

    client, err := ethclient.Dial(config.RPCURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to RPC: %w", err)
    }

    rps := config.RequestsPerSec
    if rps <= 0 {
        rps = 10 // Default
    }

    return &Client{
        Client:      client,
        rateLimiter: rate.NewLimiter(rate.Limit(rps), rps),
        config:      config,
    }, nil
}

// NewClientSimple creates a client with default configuration
func NewClientSimple(rpcURL string) (*Client, error) {
    return NewClient(DefaultClientConfig(rpcURL))
}

// withRateLimit wraps an operation with rate limiting and optional retry
func (c *Client) withRateLimit(ctx context.Context, op func() error) error {
    if err := c.rateLimiter.Wait(ctx); err != nil {
        return fmt.Errorf("rate limit wait failed: %w", err)
    }
    return op()
}

// withRateLimitResult wraps an operation that returns a result
func withRateLimitResult[T any](c *Client, ctx context.Context, op func() (T, error)) (T, error) {
    var zero T
    if err := c.rateLimiter.Wait(ctx); err != nil {
        return zero, fmt.Errorf("rate limit wait failed: %w", err)
    }
    return op()
}

// BlockNumber returns the current block number with rate limiting
func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
    return withRateLimitResult(c, ctx, func() (uint64, error) {
        return c.Client.BlockNumber(ctx)
    })
}

// BlockByNumber returns a block with rate limiting
func (c *Client) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
    return withRateLimitResult(c, ctx, func() (*types.Block, error) {
        return c.Client.BlockByNumber(ctx, number)
    })
}

// Add similar wrappers for other commonly used methods...
```

---

### Task 2.4: Add Configuration Validation

**Issue Reference:** Audit Report 2.4
**Estimated Time:** 3 hours
**Files:** `cmd/sequencer-reader/main.go`, `cmd/sequencer-capture/main.go`

#### Implementation

Create a new validation package:

```go
// pkg/config/validation.go
package config

import (
    "errors"
    "fmt"
    "net/url"
    "runtime"

    "arbitrum-sequencer-decoder/pkg/types"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("config validation error: %s: %s", e.Field, e.Message)
}

// ValidateConfig validates and normalizes a configuration
func ValidateConfig(cfg *types.Config) error {
    var errs []error

    // Validate RPC endpoint
    if cfg.RPCEndpoint == "" {
        errs = append(errs, &ValidationError{"RPCEndpoint", "cannot be empty"})
    } else {
        if _, err := url.Parse(cfg.RPCEndpoint); err != nil {
            errs = append(errs, &ValidationError{"RPCEndpoint", fmt.Sprintf("invalid URL: %v", err)})
        }
    }

    // Validate and normalize concurrency
    if cfg.MaxConcurrency <= 0 {
        cfg.MaxConcurrency = runtime.NumCPU()
    }
    if cfg.MaxConcurrency > 100 {
        errs = append(errs, &ValidationError{"MaxConcurrency", "value too high (max 100)"})
    }

    // Validate log level
    validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
    if cfg.LogLevel != "" && !validLogLevels[cfg.LogLevel] {
        errs = append(errs, &ValidationError{"LogLevel", "must be debug, info, warn, or error"})
    }

    // Validate private key format (don't log the value!)
    if cfg.PrivateKey != "" {
        if len(cfg.PrivateKey) != 64 && len(cfg.PrivateKey) != 66 {
            errs = append(errs, &ValidationError{"PrivateKey", "invalid length"})
        }
    }

    // Validate min profit
    if cfg.MinProfit != nil && cfg.MinProfit.Sign() < 0 {
        errs = append(errs, &ValidationError{"MinProfit", "cannot be negative"})
    }

    if len(errs) > 0 {
        return errors.Join(errs...)
    }

    return nil
}
```

Update `cmd/sequencer-reader/main.go`:

```go
func loadConfig() (*types.Config, error) {
    cfg := &types.Config{
        RPCEndpoint:    *rpcEndpoint,
        SequencerURL:   *sequencerURL,
        LogLevel:       *logLevel,
        MaxConcurrency: 10,
    }

    if *configFile != "" {
        // Load from file (implement file loading)
        log.Printf("Loading configuration from: %s", *configFile)
    }

    // Validate configuration
    if err := config.ValidateConfig(cfg); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }

    // Log configuration (excluding sensitive values)
    log.Printf("Configuration loaded: RPC=%s, LogLevel=%s, MaxConcurrency=%d",
        cfg.RPCEndpoint, cfg.LogLevel, cfg.MaxConcurrency)
    if cfg.PrivateKey != "" {
        log.Println("Private key: [CONFIGURED]")
    }

    return cfg, nil
}
```

---

### Task 2.5: Oracle Copy-on-Read

**Issue Reference:** Audit Report 2.5
**Estimated Time:** 2 hours
**File:** `pkg/oracle/pool_oracle.go`

#### Implementation

```go
// GetPool returns a copy of the pool at the given address
func (o *PoolOracle) GetPool(address common.Address) (*types.Pool, error) {
    o.mu.RLock()
    defer o.mu.RUnlock()

    pool, exists := o.pools[address]
    if !exists {
        return nil, types.ErrPoolNotFound
    }

    // Return a deep copy to prevent external modification
    return copyPool(pool), nil
}

// copyPool creates a deep copy of a Pool
func copyPool(p *types.Pool) *types.Pool {
    if p == nil {
        return nil
    }

    copy := &types.Pool{
        Address:     p.Address,
        Type:        p.Type,
        Token0:      p.Token0,
        Token1:      p.Token1,
        LastUpdated: p.LastUpdated,
    }

    // Deep copy optional Token2
    if p.Token2 != nil {
        t2 := *p.Token2
        copy.Token2 = &t2
    }

    // Deep copy big.Int fields
    if p.Fee != nil {
        copy.Fee = new(big.Int).Set(p.Fee)
    }

    // Deep copy big.Float
    if p.Liquidity != nil {
        copy.Liquidity = new(big.Float).Set(p.Liquidity)
    }

    // Deep copy extra data map
    if p.ExtraData != nil {
        copy.ExtraData = make(map[string]string, len(p.ExtraData))
        for k, v := range p.ExtraData {
            copy.ExtraData[k] = v
        }
    }

    return copy
}

// GetPoolState returns a copy of the pool state
func (o *PoolOracle) GetPoolState(address common.Address) (*types.PoolState, error) {
    o.mu.RLock()
    defer o.mu.RUnlock()

    state, exists := o.states[address]
    if !exists {
        return nil, types.ErrPoolNotFound
    }

    return copyPoolState(state), nil
}

// copyPoolState creates a deep copy of a PoolState
func copyPoolState(s *types.PoolState) *types.PoolState {
    if s == nil {
        return nil
    }

    copy := &types.PoolState{
        Price0To1:   s.Price0To1,
        Price1To0:   s.Price1To0,
        LastUpdated: s.LastUpdated,
        Tick:        s.Tick,
    }

    if s.Reserves0 != nil {
        copy.Reserves0 = new(big.Int).Set(s.Reserves0)
    }
    if s.Reserves1 != nil {
        copy.Reserves1 = new(big.Int).Set(s.Reserves1)
    }
    if s.Reserves2 != nil {
        copy.Reserves2 = new(big.Int).Set(s.Reserves2)
    }
    if s.Liquidity != nil {
        copy.Liquidity = new(big.Int).Set(s.Liquidity)
    }
    if s.ActiveLiquidity != nil {
        copy.ActiveLiquidity = new(big.Int).Set(s.ActiveLiquidity)
    }

    if s.ExtraData != nil {
        copy.ExtraData = make(map[string]string, len(s.ExtraData))
        for k, v := range s.ExtraData {
            copy.ExtraData[k] = v
        }
    }

    return copy
}
```

---

## Phase 3: Medium Priority Fixes (Priority P2)

Detailed implementations for Phase 3 and 4 issues are available upon request. Key items include:

- **3.1** Camelot V3 type assertion fix
- **3.2** Consolidate duplicate type definitions into `agents/auditing/types.go`
- **3.3** Implement Uniswap V3 tick math (reference implementation available)
- **3.4** Proper file handle management with defer
- **3.5** Context timeouts for all blockchain operations
- **3.6** Fix test import references
- **3.7** Add audit trail size limits

---

## Verification Checklist

After implementing fixes, verify:

- [ ] All tests pass: `go test ./... -v`
- [ ] No race conditions: `go test ./... -race`
- [ ] Code compiles cleanly: `go build ./...`
- [ ] Linting passes: `golangci-lint run`
- [ ] Decoder correctly parses real Arbitrum transactions
- [ ] Rate limiting prevents RPC throttling
- [ ] Configuration validation catches invalid inputs
- [ ] Pool oracle returns independent copies
- [ ] Audit trail doesn't grow unbounded

---

## Testing Commands

```bash
# Run all tests
go test ./... -v

# Run with race detection
go test ./... -race

# Run specific package tests
go test ./pkg/decoder/uniswap_v3 -v

# Run benchmarks
go test ./pkg/simulator -bench=. -benchmem

# Check code coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

*Implementation plan generated by Claude Code audit process*
