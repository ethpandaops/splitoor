package beaconchain

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// Client exposes beaconchain client
type Client interface {
	// GetValidators returns a map of validators
	GetValidators(ctx context.Context, pubkeys []string) (map[string]*Validator, error)
	// GetValidator returns a validator
	GetValidator(ctx context.Context, pubkey string) (*Validator, error)
	// GetBatchSize returns the batch size
	GetBatchSize() int
	// GetMaxRequestsPerMinute returns the max requests per minute
	GetMaxRequestsPerMinute() int
	// GetCheckInterval returns the check interval
	GetCheckInterval() time.Duration
}

type client struct {
	log                  logrus.FieldLogger
	url                  string
	apikey               string
	batchSize            int
	maxRequestsPerMinute int
	checkInterval        time.Duration
	metrics              *Metrics
}

// NewClient creates a new beaconchain instance
func NewClient(ctx context.Context, log logrus.FieldLogger, monitor string, conf *Config) (*client, error) {
	return &client{
		log:                  log.WithField("module", "beaconchain"),
		url:                  conf.Endpoint,
		apikey:               conf.APIKey,
		batchSize:            conf.BatchSize,
		maxRequestsPerMinute: conf.MaxRequestsPerMinute,
		checkInterval:        conf.CheckInterval,
		metrics:              GetMetricsInstance("splitoor_beaconchain", monitor),
	}, nil
}

func (c *client) GetValidators(ctx context.Context, pubkeys []string) (map[string]*Validator, error) {
	response, err := c.getValidators(ctx, pubkeys)
	if err != nil {
		return nil, err
	}

	if response.Status != StatusOK {
		return nil, fmt.Errorf("error response from server: %s", response.Status)
	}

	validators := make(map[string]*Validator)

	// len() for nil slices is defined as zero, so we can simplify this check
	if len(response.Data) == 0 {
		return validators, nil
	}

	for i := 0; i < len(response.Data); i++ {
		// Safe access to array elements with bounds check
		if i >= len(response.Data) {
			// This should never happen due to the loop condition, but it's a defensive check
			break
		}

		validator := &response.Data[i]

		// Only add to map if validator has a valid pubkey
		if validator != nil && validator.Pubkey != "" {
			validators[validator.Pubkey] = validator
		}
	}

	return validators, nil
}

func (c *client) GetValidator(ctx context.Context, pubkey string) (*Validator, error) {
	response, err := c.getValidator(ctx, pubkey)
	if err != nil {
		return nil, err
	}

	// Check if response is nil
	if response == nil {
		return nil, fmt.Errorf("received nil response")
	}

	if response.Status != StatusOK {
		return nil, fmt.Errorf("error response from server: %s", response.Status)
	}

	// Create a copy of the data to avoid returning a pointer to potentially unstable memory
	validator := response.Data

	return &validator, nil
}

func (c *client) GetBatchSize() int {
	return c.batchSize
}

func (c *client) GetMaxRequestsPerMinute() int {
	return c.maxRequestsPerMinute
}

func (c *client) GetCheckInterval() time.Duration {
	return c.checkInterval
}
