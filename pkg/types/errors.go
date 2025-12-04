// Package types defines common error types used throughout the arbitrum-sequencer-decoder
package types

import "errors"

// Common error variables
var (
	ErrInvalidTransaction     = errors.New("invalid transaction")
	ErrUnsupportedProtocol    = errors.New("unsupported protocol")
	ErrInsufficientLiquidity  = errors.New("insufficient liquidity")
	ErrPoolNotFound           = errors.New("pool not found")
	ErrInvalidPath            = errors.New("invalid path")
	ErrSimulationFailed       = errors.New("simulation failed")
	ErrArbitrageNotProfitable = errors.New("arbitrage not profitable after fees")
	ErrOpportunityExpired     = errors.New("opportunity expired")
	ErrValidationFailed       = errors.New("validation failed")
	ErrTransactionFailed      = errors.New("transaction failed")
)

// ErrorWithCode is an interface for errors that have associated codes
type ErrorWithCode interface {
	error
	ErrorCode() string
}

// TypedError represents an error with a specific code
type TypedError struct {
	Code    string
	Message string
	Err     error
}

func (e *TypedError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *TypedError) ErrorCode() string {
	return e.Code
}

// NewTypedError creates a new TypedError
func NewTypedError(code, message string, err error) *TypedError {
	return &TypedError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
