package cmd

import (
	"fmt"
	"github.com/neoll-labs/cosmos-block-fetcher/pkg"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/neoll-labs/cosmos-block-fetcher/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
		log.Error().Msgf("Error executing command: %v", err)

		os.Exit(1)
	}
}

func init() {
	log.Info().Msg("initializing Cosmos Block Fetcher")
	rootCmd.Flags().Uint64Var(&startHeight, "start-height", 0, "Starting block height")
	rootCmd.Flags().Uint64Var(&endHeight, "end-height", 0, "Ending block height")
	rootCmd.Flags().StringVar(&nodeURL, "node-url", "", "Cosmos RPC endpoint URL")

	rootCmd.Flags().IntVar(&parallelism, "parallelism", 5, "Number of parallel fetchers")
	rootCmd.Flags().StringVar(&outputFile, "output-file", "blocks.json", "Output JSON file path")
	rootCmd.Flags().IntVar(&retryAttempts, "retry-attempts", 3, "Number of retry attempts")
	rootCmd.Flags().DurationVar(&retryDelay, "retry-delay", time.Second, "Delay between retries")

	_ = rootCmd.MarkFlagRequired("start-height")
	_ = rootCmd.MarkFlagRequired("end-height")
	_ = rootCmd.MarkFlagRequired("node-url")

	// Set log level from environment variable or flag
	logLevel := os.Getenv("LOG_LEVEL")
	switch strings.ToLower(logLevel) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Add global fields if needed
	log.Logger = log.With().
		Str("service", "block-fetcher").
		Logger()
}

func fetchBlocks() error {
	fetchr := pkg.NewFetcher(nodeURL, retryAttempts, retryDelay)

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
		log.Info().Msgf("starting #%d work group of %d", i, parallelism)
		wg.Add(1)

		go func() {

			defer wg.Done()
			for height := range heights {
				log.Info().Msgf("fetch block #%d", height)

				block, err := fetchr.FetchBlock(height)
				if err != nil {
					log.Error().Msgf("fetch block #%d error %e", height, err)
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
				return pkg.WriteOutput(outputFile, output)
			}
			blocks = append(blocks, block)
		case err := <-errors:
			if err != nil {

				return err
			}
		}
	}
}
