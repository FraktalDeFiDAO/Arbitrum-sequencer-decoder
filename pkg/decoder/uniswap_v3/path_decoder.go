// Package uniswap_v3 provides path decoding functionality for Uniswap V3
package uniswap_v3

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

// DecodePath decodes the Uniswap V3 path bytes into token addresses and fees
// The path format is: token0, fee, token1, fee, token2, ..., tokenN
// where each token is 20 bytes and fee is 3 bytes
func DecodePath(pathBytes []byte) ([]common.Address, error) {
	if len(pathBytes) == 0 {
		return nil, errors.New("empty path")
	}

	var tokens []common.Address
	i := 0

	for i < len(pathBytes) {
		// Each token address is 20 bytes
		if i+20 > len(pathBytes) {
			return nil, errors.New("invalid path: insufficient bytes for token address")
		}

		token := common.BytesToAddress(pathBytes[i : i+20])
		tokens = append(tokens, token)
		i += 20

		// After each token except the last one, there's a 3-byte fee
		if i < len(pathBytes) {
			// Skip 3 bytes for the fee (we don't need to decode it in this context)
			if i+3 > len(pathBytes) {
				return nil, errors.New("invalid path: insufficient bytes for fee")
			}
			i += 3
		}
	}

	return tokens, nil
}

// DecodePathWithFees decodes the path and returns both tokens and fees
func DecodePathWithFees(pathBytes []byte) ([]common.Address, []uint32, error) {
	if len(pathBytes) == 0 {
		return nil, nil, errors.New("empty path")
	}

	var tokens []common.Address
	var fees []uint32
	i := 0

	for i < len(pathBytes) {
		// Each token address is 20 bytes
		if i+20 > len(pathBytes) {
			return nil, nil, errors.New("invalid path: insufficient bytes for token address")
		}

		token := common.BytesToAddress(pathBytes[i : i+20])
		tokens = append(tokens, token)
		i += 20

		// After each token except the last one, there's a 3-byte fee
		if i < len(pathBytes) {
			// Extract 3 bytes for the fee (we don't need to decode it in this context)
			if i+3 > len(pathBytes) {
				return nil, nil, errors.New("invalid path: insufficient bytes for fee")
			}

			// Convert the 3-byte fee to uint32
			fee := uint32(0)
			for j := 0; j < 3; j++ {
				fee = (fee << 8) | uint32(pathBytes[i+j])
			}
			fees = append(fees, fee)

			i += 3
		}
	}

	// The number of fees should be one less than the number of tokens
	if len(fees) != len(tokens)-1 && len(tokens) > 0 {
		return nil, nil, errors.New("invalid path: mismatch between tokens and fees count")
	}

	return tokens, fees, nil
}

// FormatPath formats token addresses and fees into Uniswap V3 path bytes
func FormatPath(tokens []common.Address, fees []uint32) ([]byte, error) {
	if len(tokens) == 0 {
		return nil, errors.New("no tokens provided")
	}

	if len(fees) != len(tokens)-1 {
		return nil, errors.New("number of fees must be one less than number of tokens")
	}

	var path []byte
	for i, token := range tokens {
		// Add the token address (20 bytes)
		path = append(path, token.Bytes()...)

		// Add the fee (3 bytes) after each token except the last one
		if i < len(fees) {
			feeBytes := make([]byte, 3)
			feeBytes[0] = byte(fees[i] >> 16)
			feeBytes[1] = byte(fees[i] >> 8)
			feeBytes[2] = byte(fees[i])
			path = append(path, feeBytes...)
		}
	}

	return path, nil
}
