// Package decoders provides convenient access to all DEX transaction decoders.
//
// This package aggregates all available decoders for decoding DEX calldata
// from raw Arbitrum sequencer transactions.
//
// Usage:
//
//	import "github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoders"
//
//	// Get all available decoders
//	allDecoders := decoders.All()
//
//	// Decode a transaction
//	for _, dec := range allDecoders {
//	    if dec.Matches(tx, toAddress) {
//	        actions, err := dec.Decode(tx, toAddress)
//	        // handle actions...
//	    }
//	}
//
// Or import specific decoders directly:
//
//	import "github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoder/uniswap_v3"
//	decoder := uniswap_v3.NewUniswapV3Decoder()
package decoders

import (
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoder/camelot_v3"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoder/kyberswap"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoder/odos"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoder/oneinch"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoder/openocean"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoder/sushiswap"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoder/uniswap_v3"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoder/universal_router"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/decoder/zerox"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
)

// All returns all available DEX decoders.
// The decoders are returned in order of priority/frequency.
func All() []types.Decoder {
	return []types.Decoder{
		uniswap_v3.NewUniswapV3Decoder(),
		universal_router.NewUniversalRouterDecoder(),
		camelot_v3.NewCamelotV3Decoder(),
		oneinch.NewOneInchDecoder(),
		openocean.NewOpenOceanDecoder(),
		odos.NewOdosDecoder(),
		sushiswap.NewSushiSwapDecoder(),
		kyberswap.NewKyberSwapDecoder(),
		zerox.NewZeroXDecoder(),
	}
}

// UniswapV3 returns a new Uniswap V3 SwapRouter decoder.
func UniswapV3() *uniswap_v3.UniswapV3Decoder {
	return uniswap_v3.NewUniswapV3Decoder()
}

// UniversalRouter returns a new Uniswap Universal Router decoder.
func UniversalRouter() *universal_router.UniversalRouterDecoder {
	return universal_router.NewUniversalRouterDecoder()
}

// CamelotV3 returns a new Camelot V3 decoder.
func CamelotV3() *camelot_v3.CamelotV3Decoder {
	return camelot_v3.NewCamelotV3Decoder()
}

// OneInch returns a new 1inch Aggregator V5 decoder.
func OneInch() *oneinch.OneInchDecoder {
	return oneinch.NewOneInchDecoder()
}

// OpenOcean returns a new OpenOcean Exchange decoder.
func OpenOcean() *openocean.OpenOceanDecoder {
	return openocean.NewOpenOceanDecoder()
}

// Odos returns a new Odos Router V2 decoder.
func Odos() *odos.OdosDecoder {
	return odos.NewOdosDecoder()
}

// SushiSwap returns a new SushiSwap RouteProcessor decoder.
func SushiSwap() *sushiswap.SushiSwapDecoder {
	return sushiswap.NewSushiSwapDecoder()
}

// KyberSwap returns a new KyberSwap Meta Aggregator Router decoder.
func KyberSwap() *kyberswap.KyberSwapDecoder {
	return kyberswap.NewKyberSwapDecoder()
}

// ZeroX returns a new 0x Exchange Proxy decoder.
func ZeroX() *zerox.ZeroXDecoder {
	return zerox.NewZeroXDecoder()
}
