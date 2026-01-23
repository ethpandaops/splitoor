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

	baseRate rate.Limit // Store original rate for restoration after 429
	rateMu   sync.Mutex // Protect rate adjustments

	chain   string
	chainMu sync.Mutex
}

const defaultRateLimit = rate.Limit(5) // 5 req/s

// NewClient creates a new Safe Transaction Service API client.
func NewClient(
	ctx context.Context,
	log logrus.FieldLogger,
	monitor string,
	conf *Config,
) (*client, error) {
	return &client{
		log:      log.WithField("module", "safe"),
		baseURL:  conf.Endpoint,
		apiKey:   conf.APIKey,
		client:   &http.Client{},
		metrics:  GetMetricsInstance("splitoor_safe", monitor),
		limiter:  rate.NewLimiter(defaultRateLimit, 1),
		baseRate: defaultRateLimit,
	}, nil
}

// SetChain sets the chain name for the client.
func (c *client) SetChain(chain string) {
	c.chainMu.Lock()
	defer c.chainMu.Unlock()

	c.chain = chain
}

// handleRateLimitResponse pauses the rate limiter and returns a channel that closes
// when the rate limiter is restored. This allows the caller to wait for restoration.
func (c *client) handleRateLimitResponse(resp *http.Response) <-chan struct{} {
	delay := 5 * time.Second

	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			delay = time.Duration(seconds) * time.Second
		}
	}

	c.log.WithField("delay", delay).Warn("rate limited by Safe API, pausing requests")

	c.rateMu.Lock()
	c.limiter.SetLimit(0)
	c.rateMu.Unlock()

	restored := make(chan struct{})

	go func() {
		time.Sleep(delay)

		c.rateMu.Lock()
		c.limiter.SetLimit(c.baseRate)
		c.rateMu.Unlock()

		c.log.Info("rate limiter restored")

		close(restored)
	}()

	return restored
}

// getChain returns the current chain or an error if not set.
func (c *client) getChain() (string, error) {
	c.chainMu.Lock()
	defer c.chainMu.Unlock()

	if c.chain == "" {
		return "", fmt.Errorf("chain is not set")
	}

	return c.chain, nil
}

func (c *client) GetQueuedTransactions(
	ctx context.Context,
	safeAddress string,
) (*MultisigTransactionsResponse, error) {
	chain, err := c.getChain()
	if err != nil {
		return nil, err
	}

	path := "/tx-service/:chain/api/v1/safes/:safe_address/multisig-transactions/"

	doRequest := func() (*http.Response, time.Time, error) {
		if waitErr := c.limiter.Wait(ctx); waitErr != nil {
			return nil, time.Time{}, fmt.Errorf("rate limiter: %w", waitErr)
		}

		start := time.Now()

		c.metrics.ObserveRequest("GET", c.baseURL, path, chain, safeAddress)

		url := fmt.Sprintf(
			"%s/tx-service/%s/api/v1/safes/%s/multisig-transactions/?executed=false",
			c.baseURL,
			chain,
			safeAddress,
		)

		req, reqErr := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
		if reqErr != nil {
			return nil, start, fmt.Errorf("failed to create request: %w", reqErr)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, doErr := c.client.Do(req)
		if doErr != nil {
			c.metrics.ObserveResponse("GET", c.baseURL, path, "error", chain, safeAddress, time.Since(start))

			return nil, start, fmt.Errorf("failed to execute request: %w", doErr)
		}

		return resp, start, nil
	}

	resp, start, err := doRequest()
	if err != nil {
		return nil, err
	}

	// Handle 429 rate limit response with retry
	if resp.StatusCode == http.StatusTooManyRequests {
		restored := c.handleRateLimitResponse(resp)
		resp.Body.Close()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-restored:
		}

		resp, start, err = doRequest()
		if err != nil {
			return nil, err
		}
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
	chain, err := c.getChain()
	if err != nil {
		return nil, err
	}

	path := "/tx-service/:chain/api/v1/multisig-transactions/:safe_tx_hash/"

	doRequest := func() (*http.Response, time.Time, error) {
		if waitErr := c.limiter.Wait(ctx); waitErr != nil {
			return nil, time.Time{}, fmt.Errorf("rate limiter: %w", waitErr)
		}

		start := time.Now()

		c.metrics.ObserveRequest("GET", c.baseURL, path, chain, safeTxHash)

		url := fmt.Sprintf(
			"%s/tx-service/%s/api/v1/multisig-transactions/%s/",
			c.baseURL,
			chain,
			safeTxHash,
		)

		req, reqErr := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
		if reqErr != nil {
			return nil, start, fmt.Errorf("failed to create request: %w", reqErr)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, doErr := c.client.Do(req)
		if doErr != nil {
			c.metrics.ObserveResponse("GET", c.baseURL, path, "error", chain, safeTxHash, time.Since(start))

			return nil, start, fmt.Errorf("failed to execute request: %w", doErr)
		}

		return resp, start, nil
	}

	resp, start, err := doRequest()
	if err != nil {
		return nil, err
	}

	// Handle 429 rate limit response with retry
	if resp.StatusCode == http.StatusTooManyRequests {
		restored := c.handleRateLimitResponse(resp)
		resp.Body.Close()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-restored:
		}

		resp, start, err = doRequest()
		if err != nil {
			return nil, err
		}
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
	chain, err := c.getChain()
	if err != nil {
		return nil, err
	}

	path := "/tx-service/:chain/api/v1/safes/:safe_address/"

	doRequest := func() (*http.Response, time.Time, error) {
		if waitErr := c.limiter.Wait(ctx); waitErr != nil {
			return nil, time.Time{}, fmt.Errorf("rate limiter: %w", waitErr)
		}

		start := time.Now()

		c.metrics.ObserveRequest("GET", c.baseURL, path, chain, safeAddress)

		url := fmt.Sprintf(
			"%s/tx-service/%s/api/v1/safes/%s/",
			c.baseURL,
			chain,
			safeAddress,
		)

		req, reqErr := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
		if reqErr != nil {
			return nil, start, fmt.Errorf("failed to create request: %w", reqErr)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, doErr := c.client.Do(req)
		if doErr != nil {
			c.metrics.ObserveResponse("GET", c.baseURL, path, "error", chain, safeAddress, time.Since(start))

			return nil, start, fmt.Errorf("failed to execute request: %w", doErr)
		}

		return resp, start, nil
	}

	resp, start, err := doRequest()
	if err != nil {
		return nil, err
	}

	// Handle 429 rate limit response with retry
	if resp.StatusCode == http.StatusTooManyRequests {
		restored := c.handleRateLimitResponse(resp)
		resp.Body.Close()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-restored:
		}

		resp, start, err = doRequest()
		if err != nil {
			return nil, err
		}
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
