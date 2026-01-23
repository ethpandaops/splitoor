package beaconchain

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Extended tests focusing on nil handling and edge cases

func TestGetValidatorsSafetyChecks(t *testing.T) {
	// Create a test server that returns a nil Data field
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[[]Validator]{
			Status: "OK",
			Data:   nil, // Intentionally nil
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create client with test server URL
	client, err := NewClient(context.Background(), log, "test", &Config{
		Endpoint:             server.URL,
		APIKey:               "",
		BatchSize:            100,
		MaxRequestsPerMinute: 100,
		CheckInterval:        time.Second,
	})
	require.NoError(t, err)

	// Test with a nil Data field in response
	validators, err := client.GetValidators(context.Background(), []string{"0x123"})
	assert.NoError(t, err, "Should handle nil Data field without error")
	assert.Empty(t, validators, "Should return empty map when Data is nil")

	// Create a test server that returns a response with empty data array
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[[]Validator]{
			Status: "OK",
			Data:   []Validator{}, // Empty but not nil
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server2.Close()

	// Update client with new server URL
	client.url = server2.URL

	// Test with empty Data array
	validators, err = client.GetValidators(context.Background(), []string{"0x123"})
	assert.NoError(t, err, "Should handle empty Data array without error")
	assert.Empty(t, validators, "Should return empty map when Data is empty")

	// Create a test server that returns validators with nil entries
	server3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is invalid JSON but we're testing robustness
		_, _ = w.Write([]byte(`{"status":"OK","data":[null,{"pubkey":"0x123","status":"active"},null]}`))
	}))
	defer server3.Close()

	// Update client with new server URL
	client.url = server3.URL

	// Test with Data array containing nil entries
	validators, err = client.GetValidators(context.Background(), []string{"0x123"})
	assert.NoError(t, err, "Should handle nil entries in Data array")
	assert.Len(t, validators, 1, "Should filter out nil validators")
	assert.Contains(t, validators, "0x123", "Should contain the valid validator")
}

func TestGetValidatorSafetyChecks(t *testing.T) {
	// Create a test server that returns a nil response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create client with test server URL
	client, err := NewClient(context.Background(), log, "test", &Config{
		Endpoint:             server.URL,
		APIKey:               "",
		BatchSize:            100,
		MaxRequestsPerMinute: 100,
		CheckInterval:        time.Second,
	})
	require.NoError(t, err)

	// Test with a malformed response that can't be unmarshaled
	_, err = client.GetValidator(context.Background(), "0x123")
	assert.Error(t, err, "Should handle malformed response with error")

	// Create a test server that returns an error status
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[Validator]{
			Status: "ERROR",
			Data:   Validator{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server2.Close()

	// Update client with new server URL
	client.url = server2.URL

	// Test with error status
	_, err = client.GetValidator(context.Background(), "0x123")
	assert.Error(t, err, "Should handle error status with error")
	assert.Contains(t, err.Error(), "error response from server")
}

func TestRequestHandlingSafetyChecks(t *testing.T) {
	// Create a test server that returns a 500 error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create client with test server URL
	client, err := NewClient(context.Background(), log, "test", &Config{
		Endpoint:             server.URL,
		APIKey:               "",
		BatchSize:            100,
		MaxRequestsPerMinute: 100,
		CheckInterval:        time.Second,
	})
	require.NoError(t, err)

	// Test with non-200 status code
	_, err = client.GetValidator(context.Background(), "0x123")
	assert.Error(t, err, "Should handle non-200 status code with error")
	assert.Contains(t, err.Error(), "status code")

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = client.GetValidator(ctx, "0x123")
	assert.Error(t, err, "Should handle cancelled context with error")

	// Test context timeout
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	// Sleep to ensure timeout
	time.Sleep(10 * time.Millisecond)

	_, err = client.GetValidator(ctx, "0x123")
	assert.Error(t, err, "Should handle context timeout with error")
}

// Tests for the public interface methods.
func TestGetBatchSize(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	expectedBatchSize := 42

	client, err := NewClient(context.Background(), log, "test", &Config{
		Endpoint:             "http://example.com",
		APIKey:               "",
		BatchSize:            expectedBatchSize,
		MaxRequestsPerMinute: 100,
		CheckInterval:        time.Second,
	})
	require.NoError(t, err)

	assert.Equal(t, expectedBatchSize, client.GetBatchSize())
}

func TestGetMaxRequestsPerMinute(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	expectedMaxRequests := 123

	client, err := NewClient(context.Background(), log, "test", &Config{
		Endpoint:             "http://example.com",
		APIKey:               "",
		BatchSize:            100,
		MaxRequestsPerMinute: expectedMaxRequests,
		CheckInterval:        time.Second,
	})
	require.NoError(t, err)

	assert.Equal(t, expectedMaxRequests, client.GetMaxRequestsPerMinute())
}

func TestGetCheckInterval(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	expectedInterval := 42 * time.Second

	client, err := NewClient(context.Background(), log, "test", &Config{
		Endpoint:             "http://example.com",
		APIKey:               "",
		BatchSize:            100,
		MaxRequestsPerMinute: 100,
		CheckInterval:        expectedInterval,
	})
	require.NoError(t, err)

	assert.Equal(t, expectedInterval, client.GetCheckInterval())
}
