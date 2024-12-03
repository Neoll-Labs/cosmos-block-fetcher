package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
	"time"

	"github.com/neoll-labs/cosmos-block-fetcher/types"
)

type Fetcher struct {
	client        *http.Client
	nodeURL       string
	retryAttempts int
	retryDelay    time.Duration
}

func NewFetcher(nodeURL string, retryAttempts int, retryDelay time.Duration) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: time.Second * 10,
		},
		nodeURL:       nodeURL,
		retryAttempts: retryAttempts,
		retryDelay:    retryDelay,
	}
}

func (f *Fetcher) GetChainID() (string, error) {
	var status types.StatusResponse
	url := fmt.Sprintf("%s/status", f.nodeURL)

	err := f.fetchWithRetry(url, &status)
	if err != nil {
		return "", err
	}

	return status.Result.NodeInfo.Network, nil
}

func (f *Fetcher) FetchBlock(height uint64) (*types.BlockMetadata, error) {
	var blockResp types.BlockResponse
	url := fmt.Sprintf("%s/block?height=%d", f.nodeURL, height)

	err := f.fetchWithRetry(url, &blockResp)
	if err != nil {
		return nil, err
	}

	heightInt, err := strconv.ParseUint(blockResp.Result.Block.Header.Height, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse height: %w", err)
	}

	return &types.BlockMetadata{
		Height: heightInt,
		NumTxs: len(blockResp.Result.Block.Data.Txs),
	}, nil
}

func (f *Fetcher) fetchWithRetry(url string, target interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= f.retryAttempts; attempt++ {
		if attempt > 0 {
			log.Info().Msgf("fetcher sleeping - attemp #%d", attempt)
			time.Sleep(f.retryDelay * time.Duration(attempt))
		}

		resp, err := f.client.Get(url)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			continue
		}

		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return fmt.Errorf("failed after %d attempts: %w", f.retryAttempts, lastErr)
}
