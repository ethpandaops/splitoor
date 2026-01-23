package safe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// Client exposes Safe Transaction Service API client.
type Client interface {
	// GetQueuedTransactions returns queued transactions for a safe.
	GetQueuedTransactions(ctx context.Context, safeAddress string) (*MultisigTransactionsResponse, error)
	// GetTransaction returns details for a specific transaction.
	GetTransaction(ctx context.Context, safeTxHash string) (*MultisigTransaction, error)
	// GetSafe returns details for a specific safe.
	GetSafe(ctx context.Context, safeAddress string) (*SafeResponse, error)
	// SetChain sets the chain name for the client.
	SetChain(chain string)
}

type client struct {
	log     logrus.FieldLogger
	baseURL string
	apiKey  string
	client  *http.Client
	metrics *Metrics
	limiter *rate.Limiter

	chain string
	mu    sync.Mutex
}

// NewClient creates a new Safe Transaction Service API client.
func NewClient(
	ctx context.Context,
	log logrus.FieldLogger,
	monitor string,
	conf *Config,
) (*client, error) {
	return &client{
		log:     log.WithField("module", "safe"),
		baseURL: conf.Endpoint,
		apiKey:  conf.APIKey,
		client:  &http.Client{},
		metrics: GetMetricsInstance("splitoor_safe", monitor),
		limiter: rate.NewLimiter(rate.Limit(5), 1), // 5 req/s, burst of 1
	}, nil
}

// SetChain sets the chain name for the client.
func (c *client) SetChain(chain string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.chain = chain
}

func (c *client) GetQueuedTransactions(
	ctx context.Context,
	safeAddress string,
) (*MultisigTransactionsResponse, error) {
	c.mu.Lock()

	chain := c.chain
	if chain == "" {
		c.mu.Unlock()

		return nil, fmt.Errorf("chain is not set")
	}

	c.mu.Unlock()

	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter: %w", err)
	}

	path := "/tx-service/:chain/api/v1/safes/:safe_address/multisig-transactions/"
	start := time.Now()

	c.metrics.ObserveRequest("GET", c.baseURL, path, chain, safeAddress)

	url := fmt.Sprintf(
		"%s/tx-service/%s/api/v1/safes/%s/multisig-transactions/?executed=false",
		c.baseURL,
		chain,
		safeAddress,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		c.metrics.ObserveResponse("GET", c.baseURL, path, "error", chain, safeAddress, time.Since(start))

		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	c.metrics.ObserveResponse(
		"GET",
		c.baseURL,
		path,
		strconv.Itoa(resp.StatusCode),
		chain,
		safeAddress,
		time.Since(start),
	)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result MultisigTransactionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *client) GetTransaction(
	ctx context.Context,
	safeTxHash string,
) (*MultisigTransaction, error) {
	c.mu.Lock()

	chain := c.chain
	if chain == "" {
		c.mu.Unlock()

		return nil, fmt.Errorf("chain is not set")
	}

	c.mu.Unlock()

	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter: %w", err)
	}

	path := "/tx-service/:chain/api/v1/multisig-transactions/:safe_tx_hash/"
	start := time.Now()

	c.metrics.ObserveRequest("GET", c.baseURL, path, chain, safeTxHash)

	url := fmt.Sprintf(
		"%s/tx-service/%s/api/v1/multisig-transactions/%s/",
		c.baseURL,
		chain,
		safeTxHash,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		c.metrics.ObserveResponse("GET", c.baseURL, path, "error", chain, safeTxHash, time.Since(start))

		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	c.metrics.ObserveResponse(
		"GET",
		c.baseURL,
		path,
		strconv.Itoa(resp.StatusCode),
		chain,
		safeTxHash,
		time.Since(start),
	)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result MultisigTransaction
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *client) GetSafe(ctx context.Context, safeAddress string) (*SafeResponse, error) {
	c.mu.Lock()

	chain := c.chain
	if chain == "" {
		c.mu.Unlock()

		return nil, fmt.Errorf("chain is not set")
	}

	c.mu.Unlock()

	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter: %w", err)
	}

	path := "/tx-service/:chain/api/v1/safes/:safe_address/"
	start := time.Now()

	c.metrics.ObserveRequest("GET", c.baseURL, path, chain, safeAddress)

	url := fmt.Sprintf(
		"%s/tx-service/%s/api/v1/safes/%s/",
		c.baseURL,
		chain,
		safeAddress,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		c.metrics.ObserveResponse("GET", c.baseURL, path, "error", chain, safeAddress, time.Since(start))

		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	c.metrics.ObserveResponse(
		"GET",
		c.baseURL,
		path,
		strconv.Itoa(resp.StatusCode),
		chain,
		safeAddress,
		time.Since(start),
	)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result SafeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
