package safe_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethpandaops/splitoor/pkg/monitor/safe"
)

// Setup test client helper.
func setupTestClient(t *testing.T, server *httptest.Server) safe.Client {
	t.Helper()

	c, err := safe.NewClient(context.Background(), logrus.New(), "test", &safe.Config{
		Endpoint: server.URL,
		APIKey:   "test-api-key",
	})
	require.NoError(t, err)

	return c
}

func TestClient_GetQueuedTransactions(t *testing.T) {
	tests := []struct {
		name           string
		chain          string
		safeAddress    string
		serverResponse *safe.MultisigTransactionsResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name:        "success empty queue",
			chain:       "eth",
			safeAddress: "0x123",
			serverResponse: &safe.MultisigTransactionsResponse{
				Count:    0,
				Next:     nil,
				Previous: nil,
				Results:  []safe.MultisigTransaction{},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:        "success with transactions",
			chain:       "eth",
			safeAddress: "0x123",
			serverResponse: &safe.MultisigTransactionsResponse{
				Count: 1,
				Results: []safe.MultisigTransaction{
					{
						Safe:                  "0x123",
						To:                    "0x456",
						Value:                 "1000000000000000000",
						SafeTxHash:            "0xabc123",
						Nonce:                 1,
						SubmissionDate:        "2024-01-01T00:00:00Z",
						IsExecuted:            false,
						ConfirmationsRequired: 2,
						Confirmations: []safe.Confirmation{
							{
								Owner:          "0x789",
								SubmissionDate: "2024-01-01T00:00:00Z",
								SignatureType:  "EOA",
							},
						},
					},
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "missing chain",
			safeAddress:  "0x123",
			serverStatus: http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "server error",
			chain:        "eth",
			safeAddress:  "0x123",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)

				if tt.serverResponse != nil {
					err := json.NewEncoder(w).Encode(tt.serverResponse)
					require.NoError(t, err)
				}
			}))
			defer server.Close()

			c, err := safe.NewClient(context.Background(), logrus.New(), "test", &safe.Config{
				Endpoint: server.URL,
				APIKey:   "test-api-key",
			})
			require.NoError(t, err)

			if tt.chain != "" {
				c.SetChain(tt.chain)
			}

			resp, err := c.GetQueuedTransactions(context.Background(), tt.safeAddress)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.serverResponse.Count, resp.Count)

			if len(tt.serverResponse.Results) > 0 {
				assert.Equal(t, tt.serverResponse.Results[0].SafeTxHash, resp.Results[0].SafeTxHash)
			}
		})
	}
}

func TestClient_GetTransaction(t *testing.T) {
	tests := []struct {
		name           string
		chain          string
		safeTxHash     string
		serverResponse *safe.MultisigTransaction
		serverStatus   int
		wantErr        bool
	}{
		{
			name:       "success",
			chain:      "eth",
			safeTxHash: "0xabc123",
			serverResponse: &safe.MultisigTransaction{
				Safe:                  "0x123",
				To:                    "0x456",
				Value:                 "1000000000000000000",
				SafeTxHash:            "0xabc123",
				Nonce:                 1,
				SubmissionDate:        "2024-01-01T00:00:00Z",
				IsExecuted:            false,
				ConfirmationsRequired: 2,
				Confirmations: []safe.Confirmation{
					{
						Owner:          "0x789",
						SubmissionDate: "2024-01-01T00:00:00Z",
						SignatureType:  "EOA",
					},
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "missing chain",
			safeTxHash:   "0xabc123",
			serverStatus: http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "server error",
			chain:        "eth",
			safeTxHash:   "0xabc123",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)

				if tt.serverResponse != nil {
					err := json.NewEncoder(w).Encode(tt.serverResponse)
					require.NoError(t, err)
				}
			}))
			defer server.Close()

			c, err := safe.NewClient(context.Background(), logrus.New(), "test", &safe.Config{
				Endpoint: server.URL,
				APIKey:   "test-api-key",
			})
			require.NoError(t, err)

			if tt.chain != "" {
				c.SetChain(tt.chain)
			}

			tx, err := c.GetTransaction(context.Background(), tt.safeTxHash)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.serverResponse.SafeTxHash, tx.SafeTxHash)
			assert.Equal(t, tt.serverResponse.Safe, tx.Safe)
			assert.Equal(t, tt.serverResponse.IsExecuted, tx.IsExecuted)
		})
	}
}

func TestClient_SetChain(t *testing.T) {
	c, err := safe.NewClient(context.Background(), logrus.New(), "test", &safe.Config{
		Endpoint: "http://localhost:1234", // Use non-routable address to fail fast
		APIKey:   "test-api-key",
	})
	require.NoError(t, err)

	// Test concurrent chain updates
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)

		fn := func(id int) {
			defer wg.Done()

			c.SetChain(fmt.Sprintf("chain%d", id))
		}

		go fn(i)
	}

	wg.Wait()

	// Verify we can still make requests after concurrent updates
	c.SetChain("eth")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = c.GetQueuedTransactions(ctx, "0x123")
	assert.Error(t, err)
}

func TestClient_URLConstruction(t *testing.T) {
	chain := "eth"
	safeAddress := "0x123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := fmt.Sprintf("/tx-service/%s/api/v1/safes/%s/multisig-transactions/", chain, safeAddress)
		assert.Equal(t, expectedPath, r.URL.Path)
		assert.Equal(t, "executed=false", r.URL.RawQuery)

		err := json.NewEncoder(w).Encode(&safe.MultisigTransactionsResponse{})
		require.NoError(t, err)
	}))
	defer server.Close()

	c, err := safe.NewClient(context.Background(), logrus.New(), "test", &safe.Config{
		Endpoint: server.URL,
		APIKey:   "test-api-key",
	})
	require.NoError(t, err)

	c.SetChain(chain)
	_, err = c.GetQueuedTransactions(context.Background(), safeAddress)
	require.NoError(t, err)
}

func TestClient_RequestMetrics(t *testing.T) {
	tests := []struct {
		name        string
		chain       string
		path        string
		statusCode  int
		shouldError bool
	}{
		{
			name:       "success request",
			chain:      "eth",
			path:       "/tx-service/eth/api/v1/safes/0x123/multisig-transactions/",
			statusCode: http.StatusOK,
		},
		{
			name:        "error request",
			chain:       "eth",
			path:        "/tx-service/eth/api/v1/safes/0x123/multisig-transactions/",
			statusCode:  http.StatusInternalServerError,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.path, r.URL.Path)
				w.WriteHeader(tt.statusCode)
				err := json.NewEncoder(w).Encode(&safe.MultisigTransactionsResponse{})
				require.NoError(t, err)
			}))
			defer server.Close()

			c := setupTestClient(t, server)
			c.SetChain(tt.chain)

			_, err := c.GetQueuedTransactions(context.Background(), "0x123")
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_InvalidResponses(t *testing.T) {
	tests := []struct {
		name     string
		response string
		wantErr  bool
	}{
		{
			name:     "invalid json",
			response: "{invalid json",
			wantErr:  true,
		},
		{
			name:     "empty response",
			response: "",
			wantErr:  true,
		},
		{
			name:     "null response",
			response: "null",
			wantErr:  false, // null is a valid JSON response
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)

				_, err := w.Write([]byte(tt.response))
				require.NoError(t, err)
			}))
			defer server.Close()

			c := setupTestClient(t, server)
			c.SetChain("eth")

			_, err := c.GetQueuedTransactions(context.Background(), "0x123")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := setupTestClient(t, server)
	c.SetChain("eth")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.GetQueuedTransactions(ctx, "0x123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), context.DeadlineExceeded.Error())
}

func TestClient_ParallelRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(&safe.MultisigTransactionsResponse{})
		require.NoError(t, err)
	}))
	defer server.Close()

	c := setupTestClient(t, server)
	c.SetChain("eth")

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)

		fn := func() {
			defer wg.Done()

			_, err := c.GetQueuedTransactions(context.Background(), "0x123")
			assert.NoError(t, err)
		}

		go fn()
	}

	wg.Wait()
}

func TestClient_GetSafe(t *testing.T) {
	tests := []struct {
		name           string
		chain          string
		safeAddress    string
		serverResponse *safe.SafeResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name:        "success",
			chain:       "eth",
			safeAddress: "0xc31Fb5899401E804C412B74a5bfFFb2B26222F3d",
			serverResponse: &safe.SafeResponse{
				Address:   "0xc31Fb5899401E804C412B74a5bfFFb2B26222F3d",
				Nonce:     3,
				Threshold: 4,
				Owners: []string{
					"0xdead09833B4e3ac912dF77d2eAEf4F117e787811",
					"0xdeadDB4896EB07A28b75B0784CbBed8503A09e22",
					"0xdeadc4752e998B1c04B8a89Dc1F3B07E5aaf1333",
					"0xdeadE2F6Cf6c401B33CDCCF5e2E49d5eEbd24d44",
					"0xdeadd6a5d91C6dEaD25c1092F737918F0c2f5c55",
					"0xdeadCd808F23F138a33F5023a2dD19792bd5F766",
				},
				MasterCopy:      "0x29fcB43b46531BcA003ddC8FCB67FFE91900C762",
				FallbackHandler: "0xfd0732Dc9E303f09fCEf3a7388Ad10A83459Ec99",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "missing chain",
			safeAddress:  "0x123",
			serverStatus: http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "server error",
			chain:        "eth",
			safeAddress:  "0x123",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)

				if tt.serverResponse != nil {
					err := json.NewEncoder(w).Encode(tt.serverResponse)
					require.NoError(t, err)
				}
			}))
			defer server.Close()

			c, err := safe.NewClient(context.Background(), logrus.New(), "test", &safe.Config{
				Endpoint: server.URL,
				APIKey:   "test-api-key",
			})
			require.NoError(t, err)

			if tt.chain != "" {
				c.SetChain(tt.chain)
			}

			resp, err := c.GetSafe(context.Background(), tt.safeAddress)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.serverResponse.Address, resp.Address)
			assert.Equal(t, tt.serverResponse.Nonce, resp.Nonce)
			assert.Equal(t, tt.serverResponse.Threshold, resp.Threshold)
			assert.Equal(t, len(tt.serverResponse.Owners), len(resp.Owners))
		})
	}
}

func TestClient_AuthorizationHeader(t *testing.T) {
	apiKey := "test-api-key-12345"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header is set
		authHeader := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer "+apiKey, authHeader)

		err := json.NewEncoder(w).Encode(&safe.MultisigTransactionsResponse{})
		require.NoError(t, err)
	}))
	defer server.Close()

	c, err := safe.NewClient(context.Background(), logrus.New(), "test", &safe.Config{
		Endpoint: server.URL,
		APIKey:   apiKey,
	})
	require.NoError(t, err)

	c.SetChain("eth")

	_, err = c.GetQueuedTransactions(context.Background(), "0x123")
	require.NoError(t, err)
}

func TestClient_RateLimitRetry(t *testing.T) {
	var requestCount int

	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		requestCount++

		count := requestCount

		if count == 1 {
			// First request returns 429 with Retry-After header
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)

			return
		}

		// Second request succeeds
		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(&safe.MultisigTransactionsResponse{
			Count:   1,
			Results: []safe.MultisigTransaction{{SafeTxHash: "0xabc"}},
		})
		require.NoError(t, err)
	}))
	defer server.Close()

	c, err := safe.NewClient(context.Background(), logrus.New(), "test", &safe.Config{
		Endpoint: server.URL,
		APIKey:   "test-key",
	})
	require.NoError(t, err)

	c.SetChain("eth")

	resp, err := c.GetQueuedTransactions(context.Background(), "0x123")
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Count)
	assert.Equal(t, "0xabc", resp.Results[0].SafeTxHash)

	mu.Lock()
	assert.Equal(t, 2, requestCount, "expected exactly 2 requests (initial + retry)")
	mu.Unlock()
}

func TestClient_RateLimitRetry_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always return 429 with long delay
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	c, err := safe.NewClient(context.Background(), logrus.New(), "test", &safe.Config{
		Endpoint: server.URL,
		APIKey:   "test-key",
	})
	require.NoError(t, err)

	c.SetChain("eth")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = c.GetQueuedTransactions(ctx, "0x123")
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestClient_RateLimitRetry_DefaultDelay(t *testing.T) {
	// This test verifies that the default 5s delay is used when no Retry-After header is present.
	// We do this by setting a context timeout shorter than 5s and verifying the request times out.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return 429 without Retry-After header - should use default 5s delay
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	c, err := safe.NewClient(context.Background(), logrus.New(), "test", &safe.Config{
		Endpoint: server.URL,
		APIKey:   "test-key",
	})
	require.NoError(t, err)

	c.SetChain("eth")

	// Use a context with timeout shorter than the default 5s delay
	// This should fail because the default delay exceeds our timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err = c.GetQueuedTransactions(ctx, "0x123")
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
