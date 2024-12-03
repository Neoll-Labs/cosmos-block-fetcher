package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/neoll-labs/cosmos-block-fetcher/testutil"
	"github.com/neoll-labs/cosmos-block-fetcher/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseUint64(s string) uint64 {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return val
}

func TestRootCommand(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/status":
			_, _ = w.Write([]byte(testutil.MockStatusResponse("test-chain")))
		case "/block":
			height := r.URL.Query().Get("height")
			_, _ = w.Write([]byte(testutil.MockBlockResponse(
				parseUint64(height),
				1,
				"2023-01-01T00:00:00Z",
			)))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tests := []struct {
		name      string
		args      []string
		expectErr bool
		validate  func(t *testing.T, outputFile string)
	}{
		{
			name: "valid range",
			args: []string{
				"--start-height", "100",
				"--end-height", "105",
				"--node-url", server.URL,
				"--output-file", filepath.Join(tmpDir, "valid-range.json"),
				"--parallelism", "2",
			},
			expectErr: false,
			validate: func(t *testing.T, outputFile string) {
				var output types.Output
				data, err := os.ReadFile(outputFile)
				require.NoError(t, err)
				require.NoError(t, json.Unmarshal(data, &output))

				assert.Equal(t, "test-chain", output.ChainID)
				assert.Len(t, output.Blocks, 6) // 100 to 105 inclusive
				assert.Equal(t, uint64(100), output.Blocks[0].Height)
				assert.Equal(t, uint64(105), output.Blocks[5].Height)
			},
		},
		{
			name: "invalid range",
			args: []string{
				"--start-height", "105",
				"--end-height", "100",
				"--node-url", server.URL,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd.SetArgs(tt.args)
			err := rootCmd.Execute()

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, tt.args[7]) // output-file path
			}
		})
	}
}
