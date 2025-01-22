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

// Setup test client helper
func setupTestClient(t *testing.T, server *httptest.Server) safe.Client {
	t.Helper()

	c, err := safe.NewClient(context.Background(), logrus.New(), "test", &safe.Config{
		Endpoint: server.URL,
	})
	require.NoError(t, err)

	return c
}

func TestClient_GetQueuedTransactions(t *testing.T) {
	tests := []struct {
		name           string
		chainID        string
		safeAddress    string
		serverResponse *safe.QueuedTransactionsResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name:        "success empty queue",
			chainID:     "1",
			safeAddress: "0x123",
			serverResponse: &safe.QueuedTransactionsResponse{
				Count:    0,
				Next:     nil,
				Previous: nil,
				Results:  []safe.QueuedTransactionResult{},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:        "success with transactions",
			chainID:     "1",
			safeAddress: "0x123",
			serverResponse: &safe.QueuedTransactionsResponse{
				Count: 1,
				Results: []safe.QueuedTransactionResult{
					{
						Type: "TRANSACTION",
						Transaction: &safe.Transaction{
							ID:       "123",
							TxStatus: "PENDING",
							TxInfo: safe.TransactionInfo{
								Type: "TRANSFER",
								Sender: safe.AddressInfo{
									Value: "0x123",
								},
								Recipient: safe.AddressInfo{
									Value: "0x456",
								},
							},
							ExecutionInfo: safe.ExecutionInfo{
								Nonce:                  1,
								ConfirmationsRequired:  2,
								ConfirmationsSubmitted: 1,
							},
						},
					},
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "missing chain ID",
			safeAddress:  "0x123",
			serverStatus: http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "server error",
			chainID:      "1",
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
			})
			require.NoError(t, err)

			if tt.chainID != "" {
				c.SetChainID(tt.chainID)
			}

			resp, err := c.GetQueuedTransactions(context.Background(), tt.safeAddress)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.serverResponse.Count, resp.Count)

			if len(tt.serverResponse.Results) > 0 {
				assert.Equal(t, tt.serverResponse.Results[0].Transaction.ID, resp.Results[0].Transaction.ID)
			}
		})
	}
}

func TestClient_GetTransaction(t *testing.T) {
	tests := []struct {
		name           string
		chainID        string
		safeTxHash     string
		serverResponse *safe.TransactionDetails
		serverStatus   int
		wantErr        bool
	}{
		{
			name:       "success",
			chainID:    "1",
			safeTxHash: "0x123",
			serverResponse: &safe.TransactionDetails{
				SafeAddress: "0x123",
				TxID:        "123",
				TxStatus:    "SUCCESS",
				TxInfo: safe.TransactionInfo{
					Type: "TRANSFER",
					Sender: safe.AddressInfo{
						Value: "0x123",
					},
					Recipient: safe.AddressInfo{
						Value: "0x456",
					},
				},
				DetailedExecutionInfo: safe.DetailedExecutionInfo{
					Type:                  "MULTISIG",
					Nonce:                 1,
					SafeTxHash:            "0x123",
					ConfirmationsRequired: 2,
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:       "missing chain ID",
			safeTxHash: "0x123",
			wantErr:    true,
		},
		{
			name:         "server error",
			chainID:      "1",
			safeTxHash:   "0x123",
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
			})
			require.NoError(t, err)

			if tt.chainID != "" {
				c.SetChainID(tt.chainID)
			}

			tx, err := c.GetTransaction(context.Background(), tt.safeTxHash)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.serverResponse.TxID, tx.TxID)
			assert.Equal(t, tt.serverResponse.SafeAddress, tx.SafeAddress)
			assert.Equal(t, tt.serverResponse.TxStatus, tx.TxStatus)
		})
	}
}

func TestClient_SetChainID(t *testing.T) {
	c, err := safe.NewClient(context.Background(), logrus.New(), "test", &safe.Config{
		Endpoint: "http://localhost:1234", // Use non-routable address to fail fast
	})
	require.NoError(t, err)

	// Test concurrent chain ID updates
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)

		fn := func(id int) {
			defer wg.Done()
			c.SetChainID(fmt.Sprintf("%d", id))
		}

		go fn(i)
	}

	wg.Wait()

	// Verify we can still make requests after concurrent updates
	c.SetChainID("1")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = c.GetQueuedTransactions(ctx, "0x123")
	assert.Error(t, err)
}

func TestClient_URLConstruction(t *testing.T) {
	chainID := "1"
	safeAddress := "0x123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := fmt.Sprintf("/v1/chains/%s/safes/%s/transactions/queued", chainID, safeAddress)
		assert.Equal(t, expectedPath, r.URL.Path)

		err := json.NewEncoder(w).Encode(&safe.QueuedTransactionsResponse{})
		require.NoError(t, err)
	}))
	defer server.Close()

	c, err := safe.NewClient(context.Background(), logrus.New(), "test", &safe.Config{
		Endpoint: server.URL,
	})
	require.NoError(t, err)

	c.SetChainID(chainID)
	_, err = c.GetQueuedTransactions(context.Background(), safeAddress)
	require.NoError(t, err)
}

func TestClient_RequestMetrics(t *testing.T) {
	tests := []struct {
		name        string
		chainID     string
		path        string
		statusCode  int
		shouldError bool
	}{
		{
			name:       "success request",
			chainID:    "1",
			path:       "/v1/chains/1/safes/0x123/transactions/queued",
			statusCode: http.StatusOK,
		},
		{
			name:        "error request",
			chainID:     "1",
			path:        "/v1/chains/1/safes/0x123/transactions/queued",
			statusCode:  http.StatusInternalServerError,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.path, r.URL.Path)
				w.WriteHeader(tt.statusCode)
				err := json.NewEncoder(w).Encode(&safe.QueuedTransactionsResponse{})
				require.NoError(t, err)
			}))
			defer server.Close()

			c := setupTestClient(t, server)
			c.SetChainID(tt.chainID)

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
			c.SetChainID("1")

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
	c.SetChainID("1")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.GetQueuedTransactions(ctx, "0x123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), context.DeadlineExceeded.Error())
}

func TestClient_ParallelRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(&safe.QueuedTransactionsResponse{})
		require.NoError(t, err)
	}))
	defer server.Close()

	c := setupTestClient(t, server)
	c.SetChainID("1")

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
