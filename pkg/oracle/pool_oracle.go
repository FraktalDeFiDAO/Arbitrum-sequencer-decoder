// Package oracle tracks in-memory pool states for efficient access and updates
package oracle

import (
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
	"github.com/ethereum/go-ethereum/common"
)

// PoolOracle manages in-memory pool states with concurrent access protection
type PoolOracle struct {
	mu     sync.RWMutex
	pools  map[common.Address]*types.Pool
	states map[common.Address]*types.PoolState
}

// NewPoolOracle creates a new pool oracle instance
func NewPoolOracle() *PoolOracle {
	return &PoolOracle{
		pools:  make(map[common.Address]*types.Pool),
		states: make(map[common.Address]*types.PoolState),
	}
}

// GetPool returns a copy of the pool at the given address to prevent external modification
func (o *PoolOracle) GetPool(address common.Address) (*types.Pool, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	pool, exists := o.pools[address]
	if !exists {
		return nil, types.ErrPoolNotFound
	}

	// Return a deep copy to prevent external modification of internal state
	return copyPool(pool), nil
}

// copyPool creates a deep copy of a Pool
func copyPool(p *types.Pool) *types.Pool {
	if p == nil {
		return nil
	}

	poolCopy := &types.Pool{
		Address:     p.Address,
		Type:        p.Type,
		Token0:      p.Token0,
		Token1:      p.Token1,
		LastUpdated: p.LastUpdated,
	}

	// Deep copy optional Token2
	if p.Token2 != nil {
		t2 := *p.Token2
		poolCopy.Token2 = &t2
	}

	// Deep copy big.Int fields
	if p.Fee != nil {
		poolCopy.Fee = new(big.Int).Set(p.Fee)
	}

	// Deep copy big.Float
	if p.Liquidity != nil {
		poolCopy.Liquidity = new(big.Float).Set(p.Liquidity)
	}

	// Deep copy extra data map
	if p.ExtraData != nil {
		poolCopy.ExtraData = make(map[string]string, len(p.ExtraData))
		for k, v := range p.ExtraData {
			poolCopy.ExtraData[k] = v
		}
	}

	return poolCopy
}

// UpdatePools updates pool states with new data
func (o *PoolOracle) UpdatePools(pools []*types.Pool) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	for _, pool := range pools {
		o.pools[pool.Address] = pool
	}

	return nil
}

// GetPoolState returns a copy of the detailed state of a pool to prevent external modification
func (o *PoolOracle) GetPoolState(address common.Address) (*types.PoolState, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	state, exists := o.states[address]
	if !exists {
		return nil, types.ErrPoolNotFound
	}

	// Return a deep copy to prevent external modification of internal state
	return copyPoolState(state), nil
}

// copyPoolState creates a deep copy of a PoolState
func copyPoolState(s *types.PoolState) *types.PoolState {
	if s == nil {
		return nil
	}

	stateCopy := &types.PoolState{
		Price0To1:   s.Price0To1,
		Price1To0:   s.Price1To0,
		LastUpdated: s.LastUpdated,
		Tick:        s.Tick,
	}

	// Deep copy big.Int fields
	if s.Reserves0 != nil {
		stateCopy.Reserves0 = new(big.Int).Set(s.Reserves0)
	}
	if s.Reserves1 != nil {
		stateCopy.Reserves1 = new(big.Int).Set(s.Reserves1)
	}
	if s.Reserves2 != nil {
		stateCopy.Reserves2 = new(big.Int).Set(s.Reserves2)
	}
	if s.Liquidity != nil {
		stateCopy.Liquidity = new(big.Int).Set(s.Liquidity)
	}
	if s.ActiveLiquidity != nil {
		stateCopy.ActiveLiquidity = new(big.Int).Set(s.ActiveLiquidity)
	}

	// Deep copy extra data map
	if s.ExtraData != nil {
		stateCopy.ExtraData = make(map[string]string, len(s.ExtraData))
		for k, v := range s.ExtraData {
			stateCopy.ExtraData[k] = v
		}
	}

	return stateCopy
}

// UpdatePoolState updates the state of a specific pool
// A deep copy is stored to prevent external modification of internal state
func (o *PoolOracle) UpdatePoolState(address common.Address, state *types.PoolState) error {
	if state == nil {
		return errors.New("state cannot be nil")
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	// Store a deep copy to prevent external modification
	stateCopy := copyPoolState(state)
	stateCopy.LastUpdated = time.Now()
	o.states[address] = stateCopy

	return nil
}

// RemovePool removes a pool from the oracle
func (o *PoolOracle) RemovePool(address common.Address) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	delete(o.pools, address)
	delete(o.states, address)

	return nil
}

// GetAllPools returns all tracked pools
func (o *PoolOracle) GetAllPools() []*types.Pool {
	o.mu.RLock()
	defer o.mu.RUnlock()

	pools := make([]*types.Pool, 0, len(o.pools))
	for _, pool := range o.pools {
		pools = append(pools, pool)
	}

	return pools
}

// GetAllPoolStates returns all tracked pool states
func (o *PoolOracle) GetAllPoolStates() map[common.Address]*types.PoolState {
	o.mu.RLock()
	defer o.mu.RUnlock()

	states := make(map[common.Address]*types.PoolState, len(o.states))
	for addr, state := range o.states {
		states[addr] = state
	}

	return states
}

// IsPoolTracked checks if a pool is being tracked by the oracle
func (o *PoolOracle) IsPoolTracked(address common.Address) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()

	_, exists := o.pools[address]
	return exists
}

// GetRecentPools returns pools updated within the specified duration
func (o *PoolOracle) GetRecentPools(duration time.Duration) []*types.Pool {
	o.mu.RLock()
	defer o.mu.RUnlock()

	threshold := time.Now().Add(-duration)
	pools := make([]*types.Pool, 0)

	for _, pool := range o.pools {
		if pool.LastUpdated.After(threshold) {
			pools = append(pools, pool)
		}
	}

	return pools
}
