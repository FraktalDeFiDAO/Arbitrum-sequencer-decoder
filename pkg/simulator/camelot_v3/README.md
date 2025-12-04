# Camelot V3 Simulator

This package provides price impact simulation for Camelot V3 pools without using an EVM.

## Components

- `math.go`: Implements Camelot V3 swap calculations and price impact simulations

## Features

- Constant product formula calculations (similar to Uniswap V2 but with Camelot-specific adjustments)
- Price impact calculations
- Liquidity requirement estimations
- Pool state updates after swaps