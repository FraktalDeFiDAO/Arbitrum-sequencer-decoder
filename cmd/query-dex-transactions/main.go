// Command to query transactions from specific DEXes on Arbitrum
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/blockchain"
)

var (
	rpcEndpoint = flag.String("rpc", "https://arb1.arbitrum.io/rpc", "Arbitrum RPC endpoint")
	outputDir   = flag.String("output", "./testdata/sequencer", "Directory to output transaction files")
	dexName     = flag.String("dex", "", "DEX name to query (uniswap_v3, camelot_v3, curve, balancer, ramses, kyber)")
	limit       = flag.Int("limit", 10, "Number of transactions to fetch per contract")
)

func main() {
	flag.Parse()

	if *dexName == "" {
		log.Fatal("DEX name is required (-dex flag)")
	}

	// Validate dex name
	validDexes := []string{"uniswap_v3", "camelot_v3", "curve", "balancer", "ramses", "kyber"}
	isValid := false
	for _, validDex := range validDexes {
		if strings.ToLower(*dexName) == validDex {
			isValid = true
			break
		}
	}
	if !isValid {
		log.Fatalf("Invalid DEX name: %s. Valid options: %v", *dexName, validDexes)
	}

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

	// Run the query based on the DEX name
	ctx := context.Background()

	switch strings.ToLower(*dexName) {
	case "uniswap_v3":
		queryUniswapV3(ctx, client)
	case "camelot_v3":
		queryCamelotV3(ctx, client)
	case "curve":
		queryCurve(ctx, client)
	case "balancer":
		queryBalancer(ctx, client)
	case "ramses":
		queryRamses(ctx, client)
	case "kyber":
		queryKyber(ctx, client)
	}
}

func queryUniswapV3(ctx context.Context, client *blockchain.Client) {
	log.Println("Querying Uniswap V3 transactions...")

	// Get transactions for Uniswap V3 router
	transactions, err := client.QueryContractTransactions(ctx, blockchain.UniswapV3RouterAddress, *limit)
	if err != nil {
		log.Printf("Error querying Uniswap V3 router: %v", err)
	} else {
		log.Printf("Retrieved %d transactions from Uniswap V3 router", len(transactions))
		err = saveTransactions(transactions, "uniswap_v3_router_transactions.jsonl")
		if err != nil {
			log.Printf("Error saving Uniswap V3 router transactions: %v", err)
		}
	}

	// Get transactions for Uniswap V3 factory
	transactions, err = client.QueryContractTransactions(ctx, blockchain.UniswapV3FactoryAddress, *limit)
	if err != nil {
		log.Printf("Error querying Uniswap V3 factory: %v", err)
	} else {
		log.Printf("Retrieved %d transactions from Uniswap V3 factory", len(transactions))
		err = saveTransactions(transactions, "uniswap_v3_factory_transactions.jsonl")
		if err != nil {
			log.Printf("Error saving Uniswap V3 factory transactions: %v", err)
		}
	}
}

func queryCamelotV3(ctx context.Context, client *blockchain.Client) {
	log.Println("Querying Camelot V3 transactions...")

	// Get transactions for Camelot V3 router
	transactions, err := client.QueryContractTransactions(ctx, blockchain.CamelotV3RouterAddress, *limit)
	if err != nil {
		log.Printf("Error querying Camelot V3 router: %v", err)
	} else {
		log.Printf("Retrieved %d transactions from Camelot V3 router", len(transactions))
		err = saveTransactions(transactions, "camelot_v3_router_transactions.jsonl")
		if err != nil {
			log.Printf("Error saving Camelot V3 router transactions: %v", err)
		}
	}

	// Get transactions for Camelot V3 factory
	transactions, err = client.QueryContractTransactions(ctx, blockchain.CamelotV3FactoryAddress, *limit)
	if err != nil {
		log.Printf("Error querying Camelot V3 factory: %v", err)
	} else {
		log.Printf("Retrieved %d transactions from Camelot V3 factory", len(transactions))
		err = saveTransactions(transactions, "camelot_v3_factory_transactions.jsonl")
		if err != nil {
			log.Printf("Error saving Camelot V3 factory transactions: %v", err)
		}
	}
}

func queryCurve(ctx context.Context, client *blockchain.Client) {
	log.Println("Querying Curve transactions...")

	// Get transactions for Curve factory (placeholder address)
	transactions, err := client.QueryContractTransactions(ctx, blockchain.CurveFactoryAddress, *limit)
	if err != nil {
		log.Printf("Error querying Curve contracts: %v", err)
	} else {
		log.Printf("Retrieved %d transactions from Curve contracts", len(transactions))
		err = saveTransactions(transactions, "curve_transactions.jsonl")
		if err != nil {
			log.Printf("Error saving Curve transactions: %v", err)
		}
	}
}

func queryBalancer(ctx context.Context, client *blockchain.Client) {
	log.Println("Querying Balancer transactions...")

	// Get transactions for Balancer vault
	transactions, err := client.QueryContractTransactions(ctx, blockchain.BalancerVaultAddress, *limit)
	if err != nil {
		log.Printf("Error querying Balancer vault: %v", err)
	} else {
		log.Printf("Retrieved %d transactions from Balancer vault", len(transactions))
		err = saveTransactions(transactions, "balancer_transactions.jsonl")
		if err != nil {
			log.Printf("Error saving Balancer transactions: %v", err)
		}
	}
}

func queryRamses(ctx context.Context, client *blockchain.Client) {
	log.Println("Querying Ramses transactions...")

	// Get transactions for Ramses router
	transactions, err := client.QueryContractTransactions(ctx, blockchain.RamsesV2RouterAddress, *limit)
	if err != nil {
		log.Printf("Error querying Ramses router: %v", err)
	} else {
		log.Printf("Retrieved %d transactions from Ramses router", len(transactions))
		err = saveTransactions(transactions, "ramses_transactions.jsonl")
		if err != nil {
			log.Printf("Error saving Ramses transactions: %v", err)
		}
	}
}

func queryKyber(ctx context.Context, client *blockchain.Client) {
	log.Println("Querying Kyber transactions...")

	// Get transactions for Kyber router
	transactions, err := client.QueryContractTransactions(ctx, blockchain.KyberElasticRouterAddress, *limit)
	if err != nil {
		log.Printf("Error querying Kyber router: %v", err)
	} else {
		log.Printf("Retrieved %d transactions from Kyber router", len(transactions))
		err = saveTransactions(transactions, "kyber_transactions.jsonl")
		if err != nil {
			log.Printf("Error saving Kyber transactions: %v", err)
		}
	}
}

// saveTransactions saves the transactions to a JSONL file
func saveTransactions(transactions []blockchain.Transaction, filename string) error {
	// Create the file
	file, err := os.Create(*outputDir + "/" + filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write each transaction as a JSON line
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
		}

		if tx.To != nil {
			txData["to"] = tx.To.Hex()
		}

		if tx.Receipt != nil {
			txData["gas_used"] = tx.Receipt.GasUsed
			txData["status"] = tx.Receipt.Status
		}

		// Marshal to JSON
		jsonBytes, err := json.Marshal(txData)
		if err != nil {
			return fmt.Errorf("failed to marshal transaction: %w", err)
		}

		// Write with a newline
		if _, err := file.Write(append(jsonBytes, '\n')); err != nil {
			return fmt.Errorf("failed to write transaction: %w", err)
		}
	}

	log.Printf("Saved %d transactions to %s", len(transactions), *outputDir+"/"+filename)
	return nil
}
