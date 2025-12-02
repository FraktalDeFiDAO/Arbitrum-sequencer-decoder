# Building and Testing

## Purpose
Standardize how to build, test, and validate changes to the arbitrage engine. All workflows must be reproducible, fast, and compatible with CI.

## Core Requirements
- **Go version**: 1.25+
- **Dependencies**: Only `go-ethereum` (lite), stdlib, and embedded ABIs
- **No Docker required** for core logic (optional for replay servers)

## Build
Compile the sequencer reader and executor:
```bash
go build -tags=arbitrum -o bin/arb-engine ./cmd/...
```

## Testing Strategy
- **Unit tests**: Test individual decoders and simulators in isolation
- **Integration tests**: Validate full transaction flow using sequencer replay
- **Performance tests**: Profile decoder speed against live transaction volumes
- **Regression tests**: Ensure fixes don't break existing functionality

## Test Commands
```bash
# All tests
go test ./... -v

# Specific decoder
go test ./pkg/decoder -run TestUniswapV3Decoder -v

# Performance benchmark
go test ./pkg/simulator -bench=. -benchmem

# Integration tests with replay data
go test ./pkg/replay -run TestReplayIntegration -v
```

## CI/CD Pipeline
1. **Build validation**: Ensure all components compile with specified Go version
2. **Dependency check**: Verify no unauthorized external dependencies
3. **Unit test coverage**: Minimum 80% coverage across all packages
4. **Integration validation**: Test against sample sequencer data
5. **Performance regression**: Benchmarks must not degrade >10%

## Development Workflow
- Create feature branch from `develop`
- Add tests for new functionality
- Run `go fmt` and `golangci-lint` before committing
- Submit PR with performance impact notes for benchmark-heavy changes