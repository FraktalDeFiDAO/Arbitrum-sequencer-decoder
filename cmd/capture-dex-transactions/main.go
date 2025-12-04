// Command to capture real-time transactions from specific DEX contracts
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/blockchain"
	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/classifier"
)

var (
	rpcEndpoint = flag.String("rpc", "https://arb1.arbitrum.io/rpc", "Arbitrum RPC endpoint")
	outputDir   = flag.String("output", "./testdata/sequencer", "Directory to output transaction files")
	duration    = flag.Duration("duration", 30*time.Second, "Duration to capture transactions")
)

func main() {
	flag.Parse()

	// Connect to the blockchain client
	client, err := ethclient.Dial(*rpcEndpoint)
	if err != nil {
		log.Fatalf("Failed to connect to RPC: %v", err)
	}
	defer client.Close()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Run the capture
	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	log.Println("Starting to capture transactions from DEX contracts...")

	// Get the latest block number to start from
	latestBlock, err := client.BlockNumber(ctx)
	if err != nil {
		log.Fatalf("Failed to get latest block number: %v", err)
	}

	log.Printf("Starting capture from block: %d", latestBlock)

	// Wait for some new blocks to be mined during the capture period
	select {
	case <-ctx.Done():
		log.Println("Capture completed")
	}
}

// TransactionCapturer handles capturing and storing transactions from DEX contracts
type TransactionCapturer struct {
	client     *ethclient.Client
	outputFile *os.File
}

// CaptureTransactionsFromDEXes captures transactions for specific DEX contracts
func CaptureTransactionsFromDEXes(ctx context.Context, client *ethclient.Client, duration time.Duration) error {
	// Get all DEX info
	dexes := blockchain.GetAllDexes()

	log.Println("Monitoring DEX addresses:")
	for _, dex := range dexes {
		log.Printf("  %s: Router=%s, Factory=%s", dex.Name, dex.Router.Hex(), dex.Factory.Hex())
	}

	// Get the starting block number
	startBlock, err := client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get start block: %w", err)
	}

	log.Printf("Starting from block: %d", startBlock)

	// Create a context with timeout
	captureCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	// Continuously check for new blocks
	lastBlock := startBlock
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-captureCtx.Done():
			log.Println("Capture duration reached")
			return nil
		case <-ticker.C:
			currentBlock, err := client.BlockNumber(captureCtx)
			if err != nil {
				log.Printf("Error getting current block: %v", err)
				continue
			}

			if currentBlock > lastBlock {
				log.Printf("Processing %d new blocks (%d to %d)", currentBlock-lastBlock, lastBlock+1, currentBlock)

				// Process all new blocks
				for blockNum := lastBlock + 1; blockNum <= currentBlock; blockNum++ {
					if err := processBlockForDEXes(captureCtx, client, blockNum, dexes); err != nil {
						log.Printf("Error processing block %d: %v", blockNum, err)
					}
				}

				lastBlock = currentBlock
			}
		}
	}
}

// processBlockForDEXes processes a block to find transactions to DEX contracts
func processBlockForDEXes(ctx context.Context, client *ethclient.Client, blockNum uint64, dexes []blockchain.DexInfo) error {
	block, err := client.BlockByNumber(ctx, new(big.Int).SetUint64(blockNum))
	if err != nil {
		return fmt.Errorf("failed to get block %d: %w", blockNum, err)
	}

	// Get addresses to monitor
	var monitoredAddresses []common.Address
	for _, dex := range dexes {
		monitoredAddresses = append(monitoredAddresses, dex.Router)
		monitoredAddresses = append(monitoredAddresses, dex.Factory)
		monitoredAddresses = append(monitoredAddresses, dex.OtherAddrs...)
	}

	// Check each transaction in the block
	for i, tx := range block.Transactions() {
		txTo := tx.To()
		if txTo == nil {
			continue // Skip contract creation transactions for now
		}

		// Check if transaction is to a monitored address
		isMonitored := false
		var dexName string
		for _, addr := range monitoredAddresses {
			if *txTo == addr {
				isMonitored = true

				// Identify which DEX this is for
				for _, dex := range dexes {
					if *txTo == dex.Router || *txTo == dex.Factory {
						dexName = strings.ToLower(dex.Name)
						break
					}
				}
				break
			}
		}

		if isMonitored {
			// Get receipt for additional info
			receipt, err := client.TransactionReceipt(ctx, tx.Hash())
			if err != nil {
				log.Printf("Could not get receipt for DEX transaction %s: %v", tx.Hash().Hex(), err)
			}

			// Get sender
			from, err := types.Sender(types.NewLondonSigner(tx.ChainId()), tx)
			if err != nil {
				log.Printf("Could not get sender for transaction %s: %v", tx.Hash().Hex(), err)
			}

			log.Printf("Found %s transaction: %s (block %d, index %d)", dexName, tx.Hash().Hex(), blockNum, i)

			// Determine transaction type based on calldata
			txType := classifier.GetFunctionType(tx.Data())
			if txType == "" {
				txType = "unknown"
			}

			log.Printf("  Type: %s", txType)
			log.Printf("  From: %s", from.Hex())
			log.Printf("  Value: %s", tx.Value().String())
			log.Printf("  Gas: %d", tx.Gas())

			// Save the transaction details
			if err := saveTransactionDetails(tx, receipt, from, blockNum, uint(i), dexName, txType); err != nil {
				log.Printf("Error saving transaction %s: %v", tx.Hash().Hex(), err)
			}
		}
	}

	return nil
}

// saveTransactionDetails saves transaction details to an appropriate file
func saveTransactionDetails(tx *types.Transaction, receipt *types.Receipt, from common.Address,
	blockNum uint64, txIndex uint, dexName, txType string) error {

	// Create filename based on DEX name
	filename := fmt.Sprintf("%s_transactions.jsonl", dexName)
	filepath := fmt.Sprintf("%s/%s", *outputDir, filename)

	// Open file in append mode
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filepath, err)
	}
	defer file.Close()

	// Create transaction data
	txData := map[string]interface{}{
		"hash":         tx.Hash().Hex(),
		"block_number": blockNum,
		"index":        txIndex,
		"from":         from.Hex(),
		"to":           tx.To().Hex(),
		"value":        tx.Value().String(),
		"input":        fmt.Sprintf("0x%x", tx.Data()),
		"gas":          tx.Gas(),
		"gas_price":    tx.GasPrice().String(),
		"nonce":        tx.Nonce(),
		"type":         tx.Type(),
		"dex_name":     dexName,
		"tx_type":      txType,
	}

	if receipt != nil {
		txData["gas_used"] = receipt.GasUsed
		txData["status"] = receipt.Status
	}

	// Marshal to JSON
	jsonBytes := []byte(fmt.Sprintf("%+v\n", txData))

	// Write to file
	if _, err := file.Write(jsonBytes); err != nil {
		return fmt.Errorf("failed to write transaction to file: %w", err)
	}

	return nil
}
