// Script to query historical transactions for all supported DEXes
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/blockchain"
)

var (
	rpcEndpoint = flag.String("rpc", "https://arb1.arbitrum.io/rpc", "Arbitrum RPC endpoint")
	outputDir   = flag.String("output", "./testdata/sequencer", "Directory to output transaction files")
	limit       = flag.Int("limit", 10, "Number of transactions to fetch per contract")
)

func main() {
	flag.Parse()

	// Connect to the blockchain client
	client, err := blockchain.NewClient(*rpcEndpoint)
	if err != nil {
		log.Fatalf("Failed to connect to RPC: %v", err)
	}
	defer client.Close()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	ctx := context.Background()

	// Get the latest block number to calculate a historical range
	latestBlock, err := client.BlockNumber(ctx)
	if err != nil {
		log.Fatalf("Failed to get latest block: %v", err)
	}

	log.Printf("Latest block: %d", latestBlock)

	// Go back 200,000 blocks for a good historical range (approximately 1-2 weeks on Arbitrum)
	fromBlockNum := uint64(0)
	if latestBlock > 200000 {
		fromBlockNum = latestBlock - 200000
	}

	fromBlock := new(big.Int).SetUint64(fromBlockNum)
	toBlock := new(big.Int).SetUint64(latestBlock)

	log.Printf("Querying blocks %d to %d", fromBlockNum, latestBlock)

	// Get all DEXes to query
	dexes := blockchain.GetAllDexes()

	// Query each DEX
	for _, dex := range dexes {
		log.Printf("Querying %s (router: %s, factory: %s)", dex.Name, dex.Router.Hex(), dex.Factory.Hex())

		// Query router transactions
		if err := queryAndSave(ctx, client, dex.Name, "router", dex.Router, fromBlock, toBlock); err != nil {
			log.Printf("Error querying %s router: %v", dex.Name, err)
		}

		// Query factory transactions
		if err := queryAndSave(ctx, client, dex.Name, "factory", dex.Factory, fromBlock, toBlock); err != nil {
			log.Printf("Error querying %s factory: %v", dex.Name, err)
		}

		// Add a small delay between queries to be respectful to the RPC
		time.Sleep(2 * time.Second)
	}

	log.Println("Historical query completed")
}

// queryAndSave queries transactions for an address and saves them to a file
func queryAndSave(ctx context.Context, client *blockchain.Client, dexName, contractType string,
	address common.Address, fromBlock, toBlock *big.Int) error {

	filter := blockchain.ContractFilter{
		Addresses: []common.Address{address},
		FromBlock: fromBlock,
		ToBlock:   toBlock,
	}

	transactions, err := client.GetTransactionsByContract(ctx, filter, *limit)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	log.Printf("Found %d %s %s transactions", len(transactions), dexName, contractType)

	// Create filename based on DEX and contract type
	filename := fmt.Sprintf("%s_%s_transactions.jsonl", dexName, contractType)
	filepath := fmt.Sprintf("%s/%s", *outputDir, filename)

	// Create the file
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Save transactions
	for _, tx := range transactions {
		txData := map[string]interface{}{
			"hash":              tx.Hash.Hex(),
			"block_number":      tx.BlockNumber.String(),
			"block_hash":        tx.BlockHash.Hex(),
			"transaction_index": tx.Index,
			"from":              tx.From.Hex(),
			"value":             tx.Value.String(),
			"input":             fmt.Sprintf("0x%x", tx.Input),
			"gas":               tx.Gas,
			"gas_price":         tx.GasPrice.String(),
			"nonce":             tx.Nonce,
			"type":              tx.Type,
			"dex":               dexName,
			"contract_type":     contractType,
			"contract_address":  address.Hex(),
		}

		if tx.To != nil {
			txData["to"] = tx.To.Hex()
		}

		if tx.Receipt != nil {
			txData["gas_used"] = tx.Receipt.GasUsed
			txData["status"] = tx.Receipt.Status
		}

		// Write as JSON line
		if _, err := file.WriteString(fmt.Sprintf("%+v\n", txData)); err != nil {
			return fmt.Errorf("failed to write transaction: %w", err)
		}
	}

	log.Printf("Saved %d %s %s transactions to %s", len(transactions), dexName, contractType, filepath)
	return nil
}
