package safe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Client exposes Safe API client
type Client interface {
	// GetQueuedTransactions returns queued transactions for a safe
	GetQueuedTransactions(ctx context.Context, safeAddress string) (*QueuedTransactionsResponse, error)
	// GetTransaction returns details for a specific transaction
	GetTransaction(ctx context.Context, safeTxHash string) (*TransactionDetails, error)
	// GetSafe returns details for a specific safe
	GetSafe(ctx context.Context, safeAddress string) (*SafeResponse, error)
	// CheckSigners returns true if all signers are owners of the safe
	CheckSigners(ctx context.Context, safeAddress string) (bool, error)
	// SetChainID sets the chain ID for the client
	SetChainID(chainID string)
}

type client struct {
	log     logrus.FieldLogger
	baseURL string
	signers []string
	client  *http.Client
	metrics *Metrics

	chainID string
	mu      sync.Mutex
}

// NewClient creates a new Safe API client
func NewClient(ctx context.Context, log logrus.FieldLogger, monitor string, conf *Config) (*client, error) {
	return &client{
		log:     log.WithField("module", "safe"),
		baseURL: conf.Endpoint,
		signers: conf.Signers,
		client:  &http.Client{},
		metrics: GetMetricsInstance("splitoor_safe", monitor),
	}, nil
}

func (c *client) SetChainID(chainID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.chainID = chainID
}

func (c *client) GetQueuedTransactions(ctx context.Context, safeAddress string) (*QueuedTransactionsResponse, error) {
	c.mu.Lock()

	cid := c.chainID
	if cid == "" {
		c.mu.Unlock()

		return nil, fmt.Errorf("chain ID is not set")
	}

	c.mu.Unlock()

	path := "/v1/chains/:chain_id/safes/:safe_address/transactions/queued"
	start := time.Now()

	c.metrics.ObserveRequest("GET", c.baseURL, path, cid, safeAddress)

	url := fmt.Sprintf("%s/v1/chains/%s/safes/%s/transactions/queued", c.baseURL, cid, safeAddress)

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.metrics.ObserveResponse("GET", c.baseURL, path, "error", cid, safeAddress, time.Since(start))

		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	c.metrics.ObserveResponse("GET", c.baseURL, path, strconv.Itoa(resp.StatusCode), cid, safeAddress, time.Since(start))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result QueuedTransactionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *client) GetTransaction(ctx context.Context, safeTxHash string) (*TransactionDetails, error) {
	c.mu.Lock()

	cid := c.chainID
	if cid == "" {
		c.mu.Unlock()

		return nil, fmt.Errorf("chain ID is not set")
	}

	c.mu.Unlock()

	path := "/v1/chains/:chain_id/transactions/:safe_tx_hash"
	start := time.Now()

	c.metrics.ObserveRequest("GET", c.baseURL, path, cid, safeTxHash)

	url := fmt.Sprintf("%s/v1/chains/%s/transactions/%s", c.baseURL, cid, safeTxHash)

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.metrics.ObserveResponse("GET", c.baseURL, path, "error", cid, safeTxHash, time.Since(start))

		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	c.metrics.ObserveResponse("GET", c.baseURL, path, strconv.Itoa(resp.StatusCode), cid, safeTxHash, time.Since(start))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result TransactionDetails
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *client) GetSafe(ctx context.Context, safeAddress string) (*SafeResponse, error) {
	c.mu.Lock()

	cid := c.chainID
	if cid == "" {
		c.mu.Unlock()

		return nil, fmt.Errorf("chain ID is not set")
	}

	c.mu.Unlock()

	path := "/v1/chains/:chain_id/safes/:safe_address"
	start := time.Now()

	c.metrics.ObserveRequest("GET", c.baseURL, path, cid, safeAddress)

	url := fmt.Sprintf("%s/v1/chains/%s/safes/%s", c.baseURL, cid, safeAddress)

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.metrics.ObserveResponse("GET", c.baseURL, path, "error", cid, safeAddress, time.Since(start))

		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	c.metrics.ObserveResponse("GET", c.baseURL, path, strconv.Itoa(resp.StatusCode), cid, safeAddress, time.Since(start))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result SafeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *client) CheckSigners(ctx context.Context, safeAddress string) (bool, error) {
	// check if any signers are set
	if len(c.signers) == 0 {
		return false, fmt.Errorf("no signers set in config")
	}

	safe, err := c.GetSafe(ctx, safeAddress)
	if err != nil {
		return false, fmt.Errorf("failed to get safe: %w", err)
	}

	actualSigners := make(map[string]bool)
	for _, owner := range safe.Owners {
		actualSigners[strings.ToLower(owner.Value)] = true
	}

	for _, signer := range c.signers {
		if !actualSigners[strings.ToLower(signer)] {
			return false, nil
		}
	}

	return true, nil
}
