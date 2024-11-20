package testutil

import (
	"fmt"
	"strings"
)

// MockBlockResponse returns a mock block response JSON
func MockBlockResponse(height uint64, txCount int, timestamp string) string {
	return fmt.Sprintf(`{
        "result": {
            "block": {
                "header": {
                    "height": "%d",
                    "time": "%s",
                    "chain_id": "test-chain"
                },
                "data": {
                    "txs": %s
                }
            }
        }
    }`, height, timestamp, makeMockTxs(txCount))
}

// MockStatusResponse returns a mock status response JSON
func MockStatusResponse(chainID string) string {
	return fmt.Sprintf(`{
        "result": {
            "node_info": {
                "network": "%s"
            }
        }
    }`, chainID)
}

func makeMockTxs(count int) string {
	txs := make([]string, count)
	for i := 0; i < count; i++ {
		txs[i] = "\"tx\""
	}
	return "[" + strings.Join(txs, ",") + "]"
}
