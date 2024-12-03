package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/neoll-labs/cosmos-block-fetcher/types"
	"github.com/rs/zerolog/log"
	"os"
)

func WriteOutput(outputFile string, output types.Output) error {
	log.Info().Msgf("write output file #%s", outputFile)
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
