package fetcher

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/neoll-labs/cosmos-block-fetcher/testutil"
	"github.com/neoll-labs/cosmos-block-fetcher/types"
	"github.com/stretchr/testify/assert"
)

func TestFetcher_GetChainID(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func() *httptest.Server
		expected  string
		expectErr bool
	}{
		{
			name: "successful response",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/status", r.URL.Path)
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(testutil.MockStatusResponse("cosmoshub-4")))
				}))
			},
			expected:  "cosmoshub-4",
			expectErr: false,
		},
		{
			name: "server error",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			expectErr: true,
		},
		{
			name: "timeout error",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(2 * time.Second)
				}))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupMock()
			defer server.Close()

			f := NewFetcher(server.URL, 1, time.Millisecond)
			f.client.Timeout = time.Second // Set shorter timeout for tests

			chainID, err := f.GetChainID()
			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, chainID)
		})
	}
}

func TestFetcher_FetchBlock(t *testing.T) {

	tests := []struct {
		name      string
		height    uint64
		setupMock func(uint64) *httptest.Server
		expected  *types.BlockMetadata
		expectErr bool
	}{
		{
			name:   "successful fetch",
			height: 1000000,
			setupMock: func(height uint64) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					// Add mock response here
				}))
			},
			expected:  &types.BlockMetadata{Height: 1000000},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupMock(tt.height)
			defer server.Close()

			f := NewFetcher(server.URL, 1, time.Second)
			block, err := f.FetchBlock(tt.height)

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, block)
		})
	}
}
