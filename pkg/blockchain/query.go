// Package blockchain provides utilities for querying blockchain data
package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Known DEX addresses on Arbitrum
var (
	// Uniswap V3
	UniswapV3RouterAddress  = common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564")
	UniswapV3FactoryAddress = common.HexToAddress("0x1F98431c8aD98523631AE4a59f267346ea31F984")

	// Camelot V3
	CamelotV3RouterAddress  = common.HexToAddress("0xc873fEcbd354f5A56E00E70921c767647c7A5F2c")
	CamelotV3FactoryAddress = common.HexToAddress("0x6EcCab422D763aC031210895C817871473f6B3E9")

	// Curve
	CurveFactoryAddress = common.HexToAddress("0x4e3B3F2b41De373a8E49eDe8aE456e3f66A12bE4") // This is a placeholder - real Curve contracts on Arbitrum vary

	// Balancer V2
	BalancerVaultAddress = common.HexToAddress("0xBA12222222228d8Ba445958a75a0704d566BF2C8")

	// Ramses
	RamsesV2RouterAddress = common.HexToAddress("0xAAA87963EAdFc8a017b7811CE6D3b4E4b4D5Dc11")

	// Kyber
	KyberElasticRouterAddress = common.HexToAddress("0x4f9b7DEDD8865871dF65c5D3593CaCE3b7FA3349")
)

// DexInfo contains information about a DEX
type DexInfo struct {
	Name       string
	Router     common.Address
	Factory    common.Address
	OtherAddrs []common.Address
}

// GetAllDexes returns information about all supported DEXes
func GetAllDexes() []DexInfo {
	return []DexInfo{
		{
			Name:       "UniswapV3",
			Router:     UniswapV3RouterAddress,
			Factory:    UniswapV3FactoryAddress,
			OtherAddrs: []common.Address{},
		},
		{
			Name:       "CamelotV3",
			Router:     CamelotV3RouterAddress,
			Factory:    CamelotV3FactoryAddress,
			OtherAddrs: []common.Address{},
		},
		{
			Name:       "Curve",
			Router:     CurveFactoryAddress,
			Factory:    CurveFactoryAddress,
			OtherAddrs: []common.Address{},
		},
		{
			Name:       "Balancer",
			Router:     BalancerVaultAddress,
			Factory:    BalancerVaultAddress,
			OtherAddrs: []common.Address{},
		},
		{
			Name:       "Ramses",
			Router:     RamsesV2RouterAddress,
			Factory:    RamsesV2RouterAddress,
			OtherAddrs: []common.Address{},
		},
		{
			Name:       "Kyber",
			Router:     KyberElasticRouterAddress,
			Factory:    KyberElasticRouterAddress,
			OtherAddrs: []common.Address{},
		},
	}
}

// QueryDexTransactions queries transactions for a specific DEX
func (c *Client) QueryDexTransactions(ctx context.Context, dexName string, limit int) ([]Transaction, error) {
	dexes := GetAllDexes()

	var targetAddresses []common.Address
	found := false

	for _, dex := range dexes {
		if dex.Name == dexName {
			targetAddresses = append(targetAddresses, dex.Router)
			targetAddresses = append(targetAddresses, dex.Factory)
			targetAddresses = append(targetAddresses, dex.OtherAddrs...)
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("DEX %s not found", dexName)
	}

	// Get recent blocks to query
	currentBlock, err := c.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current block: %w", err)
	}

	// Query the last 100000 blocks for transactions (increased search range)
	fromBlock := new(big.Int).Sub(new(big.Int).SetUint64(currentBlock), big.NewInt(100000))
	if fromBlock.Sign() < 0 {
		fromBlock = big.NewInt(0)
	}

	filter := ContractFilter{
		Addresses: targetAddresses,
		FromBlock: fromBlock,
		ToBlock:   new(big.Int).SetUint64(currentBlock),
	}

	return c.GetTransactionsByContract(ctx, filter, limit)
}

// QueryContractTransactions queries transactions to a specific contract address
func (c *Client) QueryContractTransactions(ctx context.Context, contractAddress common.Address, limit int) ([]Transaction, error) {
	// Get recent blocks to query
	currentBlock, err := c.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current block: %w", err)
	}

	// Query the last 100000 blocks for transactions (increased search range)
	fromBlock := new(big.Int).Sub(new(big.Int).SetUint64(currentBlock), big.NewInt(100000))
	if fromBlock.Sign() < 0 {
		fromBlock = big.NewInt(0)
	}

	filter := ContractFilter{
		Addresses: []common.Address{contractAddress},
		FromBlock: fromBlock,
		ToBlock:   new(big.Int).SetUint64(currentBlock),
	}

	return c.GetTransactionsByContract(ctx, filter, limit)
}

// GetRecentBlocks gets the most recent blocks
func (c *Client) GetRecentBlocks(ctx context.Context, count int) error {
	currentBlock, err := c.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block: %w", err)
	}

	fmt.Printf("Current block: %d\n", currentBlock)

	for i := 0; i < count; i++ {
		blockNum := new(big.Int).Sub(new(big.Int).SetUint64(currentBlock), big.NewInt(int64(i)))
		if blockNum.Sign() <= 0 {
			blockNum.SetInt64(1) // Don't go below block 1
		}

		block, err := c.BlockByNumber(ctx, blockNum)
		if err != nil {
			fmt.Printf("Error fetching block %s: %v\n", blockNum.String(), err)
			continue
		}

		fmt.Printf("Block %s: %d transactions\n", blockNum.String(), len(block.Transactions()))
		time.Sleep(100 * time.Millisecond) // Small delay to prevent overwhelming the RPC
	}

	return nil
}

// GetContractCodeSize gets the size of the code at an address
func (c *Client) GetContractCodeSize(ctx context.Context, address common.Address) (int, error) {
	code, err := c.CodeAt(ctx, address, nil)
	if err != nil {
		return 0, err
	}
	return len(code), nil
}

// IsValidContract checks if an address has contract code
func (c *Client) IsValidContract(ctx context.Context, address common.Address) (bool, error) {
	codeSize, err := c.GetContractCodeSize(ctx, address)
	if err != nil {
		return false, err
	}
	return codeSize > 0, nil
}
