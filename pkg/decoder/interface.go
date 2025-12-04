// Package decoder provides interfaces and base functionality for DEX transaction decoders
package decoder

import (
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// BaseDecoder provides common functionality for all DEX decoders
type BaseDecoder struct {
	// Add any common fields that all decoders might need
	ProtocolType types.ProtocolType
	Name         string
}

// Decoder interface defines the common interface for all DEX decoders
type Decoder interface {
	// Matches checks if a transaction should be decoded by this decoder
	Matches(tx *ethtypes.Transaction, toAddress string) bool

	// Decode decodes the transaction data and returns the decoded actions
	Decode(tx *ethtypes.Transaction, toAddress string) ([]types.DecodedAction, error)

	// Protocol returns the protocol this decoder handles
	Protocol() types.ProtocolType
}

// MatchesSignature checks if the transaction calldata starts with one of the provided function signatures
func MatchesSignature(tx *ethtypes.Transaction, signatures interface{}) bool {
	calldata := tx.Data()
	if len(calldata) < 4 {
		return false
	}

	txSigBytes := calldata[:4]
	txSignature := "0x" + common.Bytes2Hex(txSigBytes)

	switch sigs := signatures.(type) {
	case []string:
		for _, sig := range sigs {
			if txSignature == sig {
				return true
			}
		}
	case [][]byte:
		for _, sig := range sigs {
			if len(sig) == 4 &&
				txSigBytes[0] == sig[0] &&
				txSigBytes[1] == sig[1] &&
				txSigBytes[2] == sig[2] &&
				txSigBytes[3] == sig[3] {
				return true
			}
		}
	}

	return false
}
