package beaconchain

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
)

func TestGetRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the request method is GET
		assert.Equal(t, "GET", r.Method)

		// Check auth header if provided
		if apiKey := r.Header.Get("apikey"); apiKey != "" {
			assert.Equal(t, "test-key", apiKey)
		}

		// Return a simple response
		w.Write([]byte(`{"status":"OK","data":{}}`))
	}))
	defer server.Close()

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create client with test server URL
	client := &client{
		log:     log,
		url:     server.URL,
		apikey:  "test-key",
		metrics: GetMetricsInstance("test", "test"),
	}

	// Test successful GET request
	data, err := client.get(context.Background(), "test", server.URL)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, `{"status":"OK","data":{}}`, string(data))
}

func TestGetRequestErrors(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create client with invalid URL to force error
	client := &client{
		log:     log,
		url:     "http://invalid-url-that-does-not-exist.example",
		metrics: GetMetricsInstance("test", "test"),
	}

	// Test network error
	_, err := client.get(context.Background(), "test", "http://invalid-url-that-does-not-exist.example")
	assert.Error(t, err)

	// Create a test server that returns a non-200 status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	// Update client URL
	client.url = server.URL

	// Test non-200 status code
	_, err = client.get(context.Background(), "test", server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status code: 400")

	// Create a server that times out
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Write([]byte(`{"status":"OK","data":{}}`))
	}))
	defer server.Close()

	// Set a short timeout and test timeout error
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	_, err = client.get(ctx, "test", server.URL)
	assert.Error(t, err)
}

func TestGetValidatorsRequest(t *testing.T) {
	validatorData := []Validator{
		{
			Pubkey: "0x123",
			Status: "active",
		},
		{
			Pubkey: "0x456",
			Status: "pending",
		},
	}

	// Create a test server for multiple validators
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the URL contains the expected pubkeys
		if strings.Contains(r.URL.Path, "0x123,0x456") {
			resp := Response[[]Validator]{
				Status: "OK",
				Data:   validatorData,
			}
			json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create client with test server URL
	client := &client{
		log:     log,
		url:     server.URL,
		metrics: GetMetricsInstance("test", "test"),
	}

	// Test getValidators with multiple pubkeys
	resp, err := client.getValidators(context.Background(), []string{"0x123", "0x456"})
	assert.NoError(t, err)
	assert.Equal(t, "OK", resp.Status)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "0x123", resp.Data[0].Pubkey)
	assert.Equal(t, "0x456", resp.Data[1].Pubkey)
}

func TestGetValidatorsSingleResponse(t *testing.T) {
	// Create a test server that returns a single validator response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a single validator response even though multiple were requested
		resp := Response[Validator]{
			Status: "OK",
			Data: Validator{
				Pubkey: "0x123",
				Status: "active",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create client with test server URL
	client := &client{
		log:     log,
		url:     server.URL,
		metrics: GetMetricsInstance("test", "test"),
	}

	// Test getValidators with response that's a single validator
	resp, err := client.getValidators(context.Background(), []string{"0x123"})
	assert.NoError(t, err)
	assert.Equal(t, "OK", resp.Status)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "0x123", resp.Data[0].Pubkey)
}

func TestGetValidatorsErrorResponse(t *testing.T) {
	// Create a test server that returns an error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[[]Validator]{
			Status: "ERROR",
			Data:   nil,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create client with test server URL
	client := &client{
		log:     log,
		url:     server.URL,
		metrics: GetMetricsInstance("test", "test"),
	}

	// GetValidators should return the error status
	_, err := client.GetValidators(context.Background(), []string{"0x123"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error response from server")
}

func TestGetValidatorRequest(t *testing.T) {
	// Create a test server for a single validator
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the URL contains the expected pubkey
		if strings.Contains(r.URL.Path, "0x123") {
			resp := Response[Validator]{
				Status: "OK",
				Data: Validator{
					Pubkey: "0x123",
					Status: "active",
				},
			}
			json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create client with test server URL
	client := &client{
		log:     log,
		url:     server.URL,
		metrics: GetMetricsInstance("test", "test"),
	}

	// Test getValidator
	resp, err := client.getValidator(context.Background(), "0x123")
	require.NoError(t, err)
	assert.Equal(t, "OK", resp.Status)
	assert.Equal(t, "0x123", resp.Data.Pubkey)
}

func TestGetValidatorRequestInvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"OK","data":invalid}`))
	}))
	defer server.Close()

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create client with test server URL
	client := &client{
		log:     log,
		url:     server.URL,
		metrics: GetMetricsInstance("test", "test"),
	}

	// Test getValidator with invalid JSON
	_, err := client.getValidator(context.Background(), "0x123")
	assert.Error(t, err)
}