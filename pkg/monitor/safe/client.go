package safe

import (
	"context"
	"encoding/json"
	"errors"
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
	// GetQueuedTransactions returns queued transactions for a safe with nonce >= minNonce.
	GetQueuedTransactions(ctx context.Context, safeAddress string, minNonce int64) (*MultisigTransactionsResponse, error)
	// GetTransaction returns details for a specific transaction.
	GetTransaction(ctx context.Context, safeTxHash string) (*MultisigTransaction, error)
	// GetSafe returns details for a specific safe.
	GetSafe(ctx context.Context, safeAddress string) (*SafeResponse, error)
	// SetChain sets the chain name for the client.
	SetChain(chain string)
}

const maxRateLimitRetries = 5

type client struct {
	log     logrus.FieldLogger
	baseURL string
	apiKey  string
	client  *http.Client
	metrics *Metrics
	limiter *rate.Limiter

	baseRate       rate.Limit    // Store original rate for restoration after 429
	rateMu         sync.Mutex    // Protect rate adjustments and backoff state
	rateLimitCh    chan struct{} // Non-nil when rate limited; closed when restored
	rateLimitCount int           // Consecutive rate limit pauses for exponential backoff

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
// when the rate limiter is restored. If already rate limited, returns the existing
// channel so multiple goroutines coalesce on the same pause window. Consecutive
// rate limits apply exponential backoff to the delay and reduce the restore rate.
func (c *client) handleRateLimitResponse(resp *http.Response) <-chan struct{} {
	c.rateMu.Lock()
	defer c.rateMu.Unlock()

	// Already rate limited, return existing channel.
	if c.rateLimitCh != nil {
		return c.rateLimitCh
	}

	c.rateLimitCount++

	delay := 5 * time.Second

	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			delay = time.Duration(seconds) * time.Second
		}
	}

	// Exponential backoff: double the delay for each consecutive rate limit, cap at 4x.
	if c.rateLimitCount > 1 {
		shift := min(c.rateLimitCount-1, 2)
		delay *= time.Duration(1 << shift)
	}

	// Reduce restore rate for consecutive rate limits: halve each time, cap at 1/4.
	shift := min(c.rateLimitCount-1, 2)
	restoreRate := c.baseRate / rate.Limit(int(1)<<shift)

	c.log.WithFields(logrus.Fields{
		"delay":       delay,
		"consecutive": c.rateLimitCount,
		"restore_rps": float64(restoreRate),
	}).Warn("rate limited by Safe API, pausing requests")

	c.limiter.SetLimit(0)

	restored := make(chan struct{})
	c.rateLimitCh = restored

	go func() {
		time.Sleep(delay)

		c.rateMu.Lock()
		c.limiter.SetLimit(restoreRate)
		c.rateLimitCh = nil
		c.rateMu.Unlock()

		c.log.WithField("rps", float64(restoreRate)).Info("rate limiter restored")

		close(restored)
	}()

	return restored
}

// resetRateLimitBackoff resets the consecutive rate limit counter and restores
// the base rate after a successful (non-429) response.
func (c *client) resetRateLimitBackoff() {
	c.rateMu.Lock()
	defer c.rateMu.Unlock()

	if c.rateLimitCount == 0 {
		return
	}

	c.rateLimitCount = 0
	c.limiter.SetLimit(c.baseRate)
}

// getChain returns the current chain or an error if not set.
func (c *client) getChain() (string, error) {
	c.chainMu.Lock()
	defer c.chainMu.Unlock()

	if c.chain == "" {
		return "", errors.New("chain is not set")
	}

	return c.chain, nil
}

func (c *client) GetQueuedTransactions(
	ctx context.Context,
	safeAddress string,
	minNonce int64,
) (*MultisigTransactionsResponse, error) {
	chain, err := c.getChain()
	if err != nil {
		return nil, err
	}

	path := "/tx-service/:chain/api/v2/safes/:safe_address/multisig-transactions/"

	doRequest := func() (*http.Response, time.Time, error) {
		if waitErr := c.limiter.Wait(ctx); waitErr != nil {
			return nil, time.Time{}, fmt.Errorf("rate limiter: %w", waitErr)
		}

		start := time.Now()

		c.metrics.ObserveRequest("GET", c.baseURL, path, chain, safeAddress)

		url := fmt.Sprintf(
			"%s/tx-service/%s/api/v2/safes/%s/multisig-transactions/?executed=false&nonce__gte=%d",
			c.baseURL,
			chain,
			safeAddress,
			minNonce,
		)

		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
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

	// Handle 429 rate limit responses with retry loop.
	for attempt := 0; resp.StatusCode == http.StatusTooManyRequests && attempt < maxRateLimitRetries; attempt++ {
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

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limit retries exhausted after %d attempts", maxRateLimitRetries)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	c.resetRateLimitBackoff()

	var result MultisigTransactionsResponse
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return nil, fmt.Errorf("failed to decode response: %w", decodeErr)
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

	path := "/tx-service/:chain/api/v2/multisig-transactions/:safe_tx_hash/"

	doRequest := func() (*http.Response, time.Time, error) {
		if waitErr := c.limiter.Wait(ctx); waitErr != nil {
			return nil, time.Time{}, fmt.Errorf("rate limiter: %w", waitErr)
		}

		start := time.Now()

		c.metrics.ObserveRequest("GET", c.baseURL, path, chain, safeTxHash)

		url := fmt.Sprintf(
			"%s/tx-service/%s/api/v2/multisig-transactions/%s/",
			c.baseURL,
			chain,
			safeTxHash,
		)

		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
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

	// Handle 429 rate limit responses with retry loop.
	for attempt := 0; resp.StatusCode == http.StatusTooManyRequests && attempt < maxRateLimitRetries; attempt++ {
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

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limit retries exhausted after %d attempts", maxRateLimitRetries)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	c.resetRateLimitBackoff()

	var result MultisigTransaction
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return nil, fmt.Errorf("failed to decode response: %w", decodeErr)
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

		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
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

	// Handle 429 rate limit responses with retry loop.
	for attempt := 0; resp.StatusCode == http.StatusTooManyRequests && attempt < maxRateLimitRetries; attempt++ {
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

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limit retries exhausted after %d attempts", maxRateLimitRetries)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	c.resetRateLimitBackoff()

	var result SafeResponse
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return nil, fmt.Errorf("failed to decode response: %w", decodeErr)
	}

	return &result, nil
}
