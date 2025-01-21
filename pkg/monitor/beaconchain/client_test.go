package beaconchain_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethpandaops/splitoor/pkg/monitor/beaconchain"
)

func TestClient_GetValidators(t *testing.T) {
	tests := []struct {
		name           string
		pubkeys        []string
		serverResponse *beaconchain.Response[[]beaconchain.Validator]
		serverStatus   int
		wantErr        bool
		expectedLen    int
	}{
		{
			name:    "success single validator",
			pubkeys: []string{"0x123"},
			serverResponse: &beaconchain.Response[[]beaconchain.Validator]{
				Status: "OK",
				Data: []beaconchain.Validator{
					{
						Pubkey:  "0x123",
						Status:  beaconchain.StatusActiveOnline,
						Balance: 32000000000,
					},
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			expectedLen:  1,
		},
		{
			name:    "success multiple validators",
			pubkeys: []string{"0x123", "0x456"},
			serverResponse: &beaconchain.Response[[]beaconchain.Validator]{
				Status: "OK",
				Data: []beaconchain.Validator{
					{
						Pubkey:  "0x123",
						Status:  beaconchain.StatusActiveOnline,
						Balance: 32000000000,
					},
					{
						Pubkey:  "0x456",
						Status:  beaconchain.StatusActiveOffline,
						Balance: 31000000000,
					},
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			expectedLen:  2,
		},
		{
			name:    "server error",
			pubkeys: []string{"0x123"},
			serverResponse: &beaconchain.Response[[]beaconchain.Validator]{
				Status: "ERROR",
			},
			serverStatus: http.StatusOK,
			wantErr:      true,
			expectedLen:  0,
		},
		{
			name:         "http error",
			pubkeys:      []string{"0x123"},
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
			expectedLen:  0,
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

			c, err := beaconchain.NewClient(context.Background(), logrus.New(), "test", &beaconchain.Config{
				Endpoint:             server.URL,
				APIKey:               "test",
				BatchSize:            100,
				MaxRequestsPerMinute: 10,
				CheckInterval:        time.Second,
			})
			require.NoError(t, err)

			validators, err := c.GetValidators(context.Background(), tt.pubkeys)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedLen, len(validators))

			if tt.serverResponse != nil {
				for _, v := range tt.serverResponse.Data {
					assert.Equal(t, v.Balance, validators[v.Pubkey].Balance)
					assert.Equal(t, v.Status, validators[v.Pubkey].Status)
				}
			}
		})
	}
}

func TestClient_GetValidator(t *testing.T) {
	tests := []struct {
		name           string
		pubkey         string
		serverResponse *beaconchain.Response[beaconchain.Validator]
		serverStatus   int
		wantErr        bool
	}{
		{
			name:   "success",
			pubkey: "0x123",
			serverResponse: &beaconchain.Response[beaconchain.Validator]{
				Status: "OK",
				Data: beaconchain.Validator{
					Pubkey:  "0x123",
					Status:  beaconchain.StatusActiveOnline,
					Balance: 32000000000,
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:   "server error",
			pubkey: "0x123",
			serverResponse: &beaconchain.Response[beaconchain.Validator]{
				Status: "ERROR",
			},
			serverStatus: http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "http error",
			pubkey:       "0x123",
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

			c, err := beaconchain.NewClient(context.Background(), logrus.New(), "test", &beaconchain.Config{
				Endpoint:             server.URL,
				APIKey:               "test",
				BatchSize:            100,
				MaxRequestsPerMinute: 10,
				CheckInterval:        time.Second,
			})
			require.NoError(t, err)

			validator, err := c.GetValidator(context.Background(), tt.pubkey)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.serverResponse.Data.Balance, validator.Balance)
			assert.Equal(t, tt.serverResponse.Data.Status, validator.Status)
		})
	}
}

func TestClient_GetConfig(t *testing.T) {
	c, err := beaconchain.NewClient(context.Background(), logrus.New(), "test", &beaconchain.Config{
		Endpoint:             "http://test",
		APIKey:               "test",
		BatchSize:            100,
		MaxRequestsPerMinute: 10,
		CheckInterval:        time.Second,
	})
	require.NoError(t, err)

	assert.Equal(t, 100, c.GetBatchSize())
	assert.Equal(t, 10, c.GetMaxRequestsPerMinute())
	assert.Equal(t, time.Second, c.GetCheckInterval())
}
func TestClient_GetValidators_EmptyPubkeys(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true

		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(&beaconchain.Response[[]beaconchain.Validator]{
			Status: "OK",
			Data:   []beaconchain.Validator{},
		})
		require.NoError(t, err)
	}))

	defer server.Close()

	c, err := beaconchain.NewClient(context.Background(), logrus.New(), "test", &beaconchain.Config{
		Endpoint:             server.URL,
		APIKey:               "test",
		BatchSize:            100,
		MaxRequestsPerMinute: 10,
		CheckInterval:        time.Second,
	})
	require.NoError(t, err)

	validators, err := c.GetValidators(context.Background(), []string{})
	require.NoError(t, err)
	assert.Empty(t, validators)
	assert.True(t, serverCalled, "server should not be called for empty pubkeys")
}

func TestClient_APIKeyHeader(t *testing.T) {
	expectedAPIKey := "test-api-key"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, expectedAPIKey, r.Header.Get("apikey"))

		err := json.NewEncoder(w).Encode(&beaconchain.Response[beaconchain.Validator]{
			Status: "OK",
			Data:   beaconchain.Validator{},
		})
		require.NoError(t, err)
	}))

	defer server.Close()

	c, err := beaconchain.NewClient(context.Background(), logrus.New(), "test", &beaconchain.Config{
		Endpoint:             server.URL,
		APIKey:               expectedAPIKey,
		BatchSize:            100,
		MaxRequestsPerMinute: 10,
		CheckInterval:        time.Second,
	})
	require.NoError(t, err)

	_, err = c.GetValidator(context.Background(), "0x123")
	require.NoError(t, err)
}

func TestClient_URLConstruction(t *testing.T) {
	pubkeys := []string{"0x123", "0x456"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, strings.HasSuffix(r.URL.Path, strings.Join(pubkeys, ",")))

		err := json.NewEncoder(w).Encode(&beaconchain.Response[[]beaconchain.Validator]{
			Status: "OK",
			Data:   []beaconchain.Validator{},
		})
		require.NoError(t, err)
	}))

	defer server.Close()

	c, err := beaconchain.NewClient(context.Background(), logrus.New(), "test", &beaconchain.Config{
		Endpoint:             server.URL,
		APIKey:               "test",
		BatchSize:            100,
		MaxRequestsPerMinute: 10,
		CheckInterval:        time.Second,
	})
	require.NoError(t, err)

	_, err = c.GetValidators(context.Background(), pubkeys)
	require.NoError(t, err)
}

func TestValidator_IsExited(t *testing.T) {
	tests := []struct {
		name     string
		status   beaconchain.Status
		expected bool
	}{
		{
			name:     "slashed validator",
			status:   beaconchain.StatusSlashed,
			expected: true,
		},
		{
			name:     "exited validator",
			status:   beaconchain.StatusExited,
			expected: true,
		},
		{
			name:     "active validator",
			status:   beaconchain.StatusActiveOnline,
			expected: false,
		},
		{
			name:     "pending validator",
			status:   beaconchain.StatusPending,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &beaconchain.Validator{Status: tt.status}
			assert.Equal(t, tt.expected, v.IsExited())
		})
	}
}
