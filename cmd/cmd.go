package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/neoll-labs/cosmos-block-fetcher/fetcher"
	"github.com/neoll-labs/cosmos-block-fetcher/types"
	"github.com/spf13/cobra"
)

var (
	startHeight   uint64
	endHeight     uint64
	nodeURL       string
	parallelism   int
	outputFile    string
	retryAttempts int
	retryDelay    time.Duration
)

var rootCmd = &cobra.Command{
	Use:   "cosmos-block-fetcher",
	Short: "A parallel block fetcher for Cosmos-based blockchains",
	Long: `Cosmos BlockMetadata Fetcher is a CLI tool that efficiently retrieves and stores 
block metadata from Cosmos-based blockchains with parallel processing and 
resilient error handling.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if endHeight < startHeight {
			return fmt.Errorf("end height must be greater than or equal to start height")
		}
		if nodeURL == "" {
			return fmt.Errorf("node URL is required")
		}

		return fetchBlocks()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().Uint64Var(&startHeight, "start-height", 0, "Starting block height")
	rootCmd.Flags().Uint64Var(&endHeight, "end-height", 0, "Ending block height")
	rootCmd.Flags().StringVar(&nodeURL, "node-url", "", "Cosmos RPC endpoint URL")

	rootCmd.Flags().IntVar(&parallelism, "parallelism", 5, "Number of parallel fetchers")
	rootCmd.Flags().StringVar(&outputFile, "output-file", "blocks.json", "Output JSON file path")
	rootCmd.Flags().IntVar(&retryAttempts, "retry-attempts", 3, "Number of retry attempts")
	rootCmd.Flags().DurationVar(&retryDelay, "retry-delay", time.Second, "Delay between retries")

	rootCmd.MarkFlagRequired("start-height")
	rootCmd.MarkFlagRequired("end-height")
	rootCmd.MarkFlagRequired("node-url")
}

func fetchBlocks() error {
	fetchr := fetcher.NewFetcher(nodeURL, retryAttempts, retryDelay)

	chainID, err := fetchr.GetChainID()
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}

	output := types.Output{
		ChainID: chainID,
		Blocks:  make([]types.BlockMetadata, 0, endHeight-startHeight+1),
	}

	results := make(chan types.BlockMetadata, parallelism)
	errors := make(chan error, parallelism)

	var wg sync.WaitGroup

	heights := make(chan uint64, parallelism)

	for i := 0; i < parallelism; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for height := range heights {
				block, err := fetchr.FetchBlock(height)
				if err != nil {
					errors <- fmt.Errorf("failed to fetch block %d: %w", height, err)
					return
				}
				results <- *block
			}
		}()
	}

	go func() {
		for height := startHeight; height <= endHeight; height++ {
			heights <- height
		}
		close(heights)
	}()

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	var blocks []types.BlockMetadata
	for {
		select {
		case block, ok := <-results:
			if !ok {
				output.Blocks = blocks
				return writeOutput(output)
			}
			blocks = append(blocks, block)
		case err := <-errors:
			if err != nil {
				return err
			}
		}
	}
}

func writeOutput(output types.Output) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
