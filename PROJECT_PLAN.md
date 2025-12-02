# Arbitrum Sequencer Decoder - MVP Project Plan for Continuous Development

## Overview
This project plan defines the MVP implementation of the Arbitrum sequencer decoder, broken down into granular tasks that can each be completed in a single context window. The system is designed to detect cross-DEX price imbalances on Arbitrum using pre-consensus sequencer data and execute atomic arbitrage opportunities. This is an MVP foundation for continuous development - as requirements emerge and new features are needed, the plan will be updated and expanded accordingly.

## Phase 1: Project Setup and Infrastructure

### Task 1.1: Create Go module and directory structure
- Initialize Go module: `go mod init arbitrum-sequencer-decoder`
- Create directory structure: `cmd/`, `pkg/decoder/`, `pkg/simulator/`, `pkg/oracle/`, `pkg/arb-engine/`, `pkg/executor/`
- Set up basic `go.mod` dependencies
- Create `go.sum` file

### Task 1.2: Create basic CLI framework
- Implement basic command line interface in `cmd/sequencer-reader/main.go`
- Set up flags for RPC endpoints, sequencer URLs
- Add basic logging framework
- Add configuration file support

### Task 1.3: Set up common types and interfaces
- Create `pkg/types/` directory
- Define common types like Address, Hash, Token, Pool, etc.
- Define interfaces for decoders, simulators, oracles
- Create common error types and constants

### Task 1.4: Implement basic transaction capture
- Create `cmd/sequencer-capture` command
- Implement connection to Arbitrum RPC endpoint
- Capture pending transactions from mempool/sequencer
- Write transactions to JSONL format in `testdata/sequencer/`

## Phase 2: Core Infrastructure

### Task 2.1: Implement transaction classifier
- Create `pkg/classifier/` package
- Implement function to identify DEX transactions by address and function signature
- Create mapping of known DEX addresses on Arbitrum
- Build function signature database for different DEX functions

### Task 2.2: Implement base decoder interface
- Define decoder interface in `pkg/decoder/interface.go`
- Create base decoder struct with common functionality
- Implement transaction matching logic
- Define decoded action types and structures

### Task 2.3: Create basic pool oracle
- Create `pkg/oracle/pool_state.go`
- Implement in-memory pool state tracking
- Create functions to update and retrieve pool states
- Add basic caching mechanisms

### Task 2.4: Set up integration testing framework
- Create integration test structure
- Set up test helpers for decoding transactions from JSONL files
- Create basic test assertions for decoder outputs
- Implement replay functionality for testing

## Phase 3: DEX Decoder Implementation (Sequential)

### Task 3.1: Create Uniswap V3 ABI bindings
- Generate Uniswap V3 router ABI bindings
- Create Go structs that match router contract functions
- Implement ABI parsing for swap functions
- Define constants for known Uniswap V3 router addresses

### Task 3.2: Implement Uniswap V3 basic decoder
- Create `pkg/decoder/uniswap_v3.go`
- Implement Matches() function to identify Uniswap V3 transactions
- Parse function selectors: `exactInput`, `exactOutput`
- Extract basic parameters: amounts, paths, recipients

### Task 3.3: Implement Uniswap V3 path decoding
- Decode Uniswap V3's compact path encoding format
- Parse token addresses and fee tiers from path bytes
- Handle multiple hop swaps in a single transaction
- Validate path structure and return errors for malformed paths

### Task 3.4: Create Uniswap V3 pool state tracking
- Implement functions to query Uniswap V3 pool states
- Fetch liquidity, current tick, and tick spacing
- Update pool state based on swap impact
- Cache frequently accessed pools

### Task 3.5: Implement Uniswap V3 price simulation
- Create `pkg/simulator/uniswap_v3_math.go`
- Implement Uniswap V3 tick math functions
- Calculate price impact for swaps without EVM
- Simulate swap outcomes for different amounts

### Task 3.6: Create Uniswap V3 decoder tests
- Write unit tests for the Uniswap V3 decoder
- Use sample transactions from `testdata/sequencer/uniswap_v3_swaps.jsonl`
- Test edge cases and error conditions
- Verify decoded output matches expected values

### Task 3.7: Create Camelot V3 ABI bindings
- Generate Camelot V3 router ABI bindings
- Create Go structs that match router contract functions
- Implement ABI parsing for swap functions
- Define constants for known Camelot V3 router addresses

### Task 3.8: Implement Camelot V3 basic decoder
- Create `pkg/decoder/camelot_v3.go`
- Implement Matches() function to identify Camelot V3 transactions
- Parse function selectors: `swapExactTokensForTokens`, etc.
- Extract basic parameters: amounts, paths, recipients

### Task 3.9: Implement Camelot V3 price simulation
- Create `pkg/simulator/camelot_v3_math.go`
- Implement constant product formula for Camelot V3
- Account for Camelot's specific fee structure (0.3%)
- Simulate swap outcomes for different amounts

### Task 3.10: Create Camelot V3 decoder tests
- Write unit tests for the Camelot V3 decoder
- Use sample transactions from `testdata/sequencer/camelot_v3_swaps.jsonl`
- Test edge cases and error conditions
- Verify decoded output matches expected values

### Task 3.11: Create Curve ABI bindings
- Generate Curve factory and pool ABI bindings
- Create Go structs that match Curve contract functions
- Implement ABI parsing for exchange functions
- Define constants for known Curve factory addresses

### Task 3.12: Implement Curve basic decoder
- Create `pkg/decoder/curve.go`
- Implement Matches() function to identify Curve transactions
- Parse function selectors: `exchange`, `exchange_underlying`, etc.
- Extract basic parameters: input/output coins, amounts

### Task 3.13: Implement Curve stableswap simulation
- Create `pkg/simulator/curve/stableswap.go`
- Implement Curve's stableswap invariant formulas
- Solve the iterative calculation for price impact
- Handle multi-asset pools (3-8 coins)

### Task 3.14: Implement Curve cryptoswap simulation
- Create `pkg/simulator/curve/cryptoswap.go`
- Implement Curve's cryptoswap invariant formulas
- Handle amplified pools with different parameters
- Calculate price impact efficiently

### Task 3.15: Create Curve decoder tests
- Write unit tests for the Curve decoder
- Use sample transactions from `testdata/sequencer/curve_swaps.jsonl`
- Test both stableswap and cryptoswap pools
- Verify decoded output matches expected values

### Task 3.16: Create Balancer V2/V3 ABI bindings
- Generate Balancer vault and pool ABI bindings
- Create Go structs that match Balancer contract functions
- Implement ABI parsing for swap and batchSwap functions
- Define constants for known Balancer vault addresses

### Task 3.17: Implement Balancer basic decoder
- Create `pkg/decoder/balancer.go`
- Implement Matches() function to identify Balancer transactions
- Parse function selectors: `swap`, `batchSwap`
- Extract basic parameters: poolId, amounts, assets

### Task 3.18: Implement Balancer weighted pool simulation
- Create `pkg/simulator/balancer/weighted.go`
- Implement Balancer's weighted pool formula
- Calculate price impact using weights
- Handle multi-asset swaps

### Task 3.19: Implement Balancer stable pool simulation
- Create `pkg/simulator/balancer/stable.go`
- Implement Balancer's stable pool formula
- Calculate price impact with amplification factors
- Handle stable swap mechanics

### Task 3.20: Create Balancer decoder tests
- Write unit tests for the Balancer decoder
- Test both weighted and stable pools
- Verify decoded output matches expected values

### Task 3.21: Create Ramses ABI bindings
- Generate Ramses router ABI bindings
- Create Go structs that match Ramses contract functions
- Implement ABI parsing for Ramses swap functions
- Define constants for known Ramses router addresses

### Task 3.22: Implement Ramses basic decoder
- Create `pkg/decoder/ramses.go`
- Implement Matches() function to identify Ramses transactions
- Parse function selectors similar to Uniswap V3
- Extract basic parameters: amounts, paths, recipients

### Task 3.23: Implement Ramses price simulation
- Create `pkg/simulator/ramses.go`
- Implement Ramses-specific pricing (similar to Uniswap V3)
- Account for Ramses additional governance features
- Calculate price impact considering veRAM effects

### Task 3.24: Create Ramses decoder tests
- Write unit tests for the Ramses decoder
- Test Ramses-specific functionality
- Verify decoded output matches expected values

### Task 3.25: Create Kyber ABI bindings
- Generate Kyber router ABI bindings
- Create Go structs that match Kyber contract functions
- Implement ABI parsing for Kyber swap functions
- Define constants for known Kyber router addresses

### Task 3.26: Implement Kyber basic decoder
- Create `pkg/decoder/kyber.go`
- Implement Matches() function to identify Kyber transactions
- Parse function selectors: classic and elastic swaps
- Extract basic parameters: amounts, paths, recipients

### Task 3.27: Implement Kyber price simulation
- Create `pkg/simulator/kyber.go`
- Implement Kyber Classic constant product pricing
- Implement Kyber Elastic concentrated liquidity pricing
- Account for dynamic fees in Classic pools

### Task 3.28: Create Kyber decoder tests
- Write unit tests for the Kyber decoder
- Test both Classic and Elastic functionality
- Verify decoded output matches expected values

## Phase 4: Arbitrage Engine Implementation

### Task 4.1: Create arbitrage opportunity types
- Define types for different arbitrage opportunities
- Create triangle arbitrage type
- Create cross-DEX arbitrage type
- Create flash swap arbitrage type

### Task 4.2: Implement pool state tracking
- Create `pkg/oracle/pool_tracker.go`
- Implement efficient caching mechanism for pool states
- Update pool states as transactions are processed
- Handle state expiration and refresh

### Task 4.3: Implement opportunity detection engine
- Create `pkg/arb-engine/opportunity_detector.go`
- Implement cross-DEX arbitrage detection
- Find price discrepancies between similar pools
- Check profitability after fees

### Task 4.4: Implement triangle arbitrage detection
- Create `pkg/arb-engine/triangle_detector.go`
- Identify triangle arbitrage opportunities
- Calculate potential profit for three-hop routes
- Verify opportunity validity

### Task 4.5: Implement flash swap arbitrage detection
- Create `pkg/arb-engine/flash_swap_detector.go`
- Identify potential flash swap opportunities
- Calculate borrowing and repayment feasibility
- Factor in flash loan fees

### Task 4.6: Create risk assessment module
- Create `pkg/arb-engine/risk_calculator.go`
- Calculate potential risks for arbitrage opportunities
- Consider gas costs, slippage, and timing
- Implement risk filtering

### Task 4.7: Implement opportunity ranking
- Create `pkg/arb-engine/ranker.go`
- Rank opportunities by expected profit
- Factor in risk, gas costs, and success probability
- Prioritize most promising opportunities

## Phase 5: Transaction Execution

### Task 5.1: Create transaction builder
- Create `pkg/executor/transaction_builder.go`
- Build arbitrage transactions from opportunities
- Encode swap calls across different DEXes
- Handle multi-hop transactions

### Task 5.2: Implement bundle submission
- Create `pkg/executor/bundle_submitter.go`
- Submit transaction bundles to relays
- Integrate with Flashbots-like services
- Handle bundle status and responses

### Task 5.3: Create gas estimation
- Create `pkg/executor/gas_estimator.go`
- Estimate gas costs for arbitrage transactions
- Account for dynamic gas prices
- Optimize for profitability

## Phase 6: Sequencer Feed Processing

### Task 6.1: Implement real-time sequencer reader
- Enhance `cmd/sequencer-reader/main.go`
- Process transactions in real-time from sequencer
- Apply decoders to identify DEX transactions
- Send to arbitrage engine for analysis

### Task 6.2: Implement transaction replay
- Create `cmd/replay/main.go`
- Replay historical sequencer data from JSONL files
- Validate decoder accuracy against known outcomes
- Benchmark performance

### Task 6.3: Implement performance monitoring
- Add metrics collection to all components
- Monitor processing latency
- Track successful arbitrage opportunities
- Create health check endpoints

## Phase 7: Testing & Validation

### Task 7.1: Integration tests
- Create end-to-end integration tests
- Test full pipeline: sequencer -> decode -> arbitrage -> execute
- Use historical transaction data
- Validate opportunity detection accuracy

### Task 7.2: Performance tests
- Benchmark decoder performance
- Test transaction processing speed
- Validate real-time processing capabilities
- Optimize bottlenecks

### Task 7.3: Load testing
- Simulate high-volume transaction processing
- Test system stability under load
- Validate memory usage and garbage collection
- Optimize for production workloads

## Phase 8: Production Deployment

### Task 8.1: Create Dockerfile
- Build minimal Docker image
- Include all necessary dependencies
- Optimize for security and size
- Multi-stage build for minimal attack surface

### Task 8.2: Create deployment configurations
- Write Kubernetes manifests
- Create configuration files for different environments
- Set up secrets management
- Define resource requirements

### Task 8.3: Create monitoring and alerting
- Set up logging system
- Create Grafana dashboards
- Implement alerting for system failures
- Monitor profitability metrics

## Task Dependencies
- Phases 1-2 must be completed before Phase 3
- Each DEX decoder in Phase 3 can be implemented in parallel after Phase 2
- Phase 4 depends on Phase 3 decoders
- Phase 5 depends on Phase 4
- Phase 6 can run in parallel with Phase 3-5
- Phases 7-8 can happen in parallel

## Continuous Development Approach
This is an MVP plan that establishes the foundational architecture and core functionality. As the system evolves and new requirements emerge:

1. Additional DEX integrations can be added following the same pattern used by existing decoders
2. New arbitrage strategies can be implemented in the arb-engine package
3. Performance optimizations can be applied as bottlenecks are identified
4. Additional risk management features can be incorporated as needed
5. New metrics and monitoring capabilities can be added as operational requirements become clearer

The modular architecture allows for continuous expansion:
- New decoders can be added to pkg/decoder/
- New simulation methods can be added to pkg/simulator/
- New arbitrage strategies can be added to pkg/arb-engine/
- New execution methods can be added to pkg/executor/

This plan will be continuously updated as new requirements, features, and capabilities are identified during development and operation of the system.