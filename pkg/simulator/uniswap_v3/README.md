# Uniswap V3 Simulator

This package provides price impact simulation for Uniswap V3 pools without using an EVM.

## Components

- `math.go`: Implements Uniswap V3 tick math and swap simulation

## Features

- Accurate Uniswap V3 pricing using concentrated liquidity math
- Tick-based price calculations
- Support for multiple fee tiers (0.01%, 0.05%, 0.3%, 1%)
- Price impact calculations
- Multi-hop swap simulation