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
	metrics              Metrics
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
		metrics:              NewMetrics("splitoor_beaconchain", monitor),
	}, nil
}

func (c *client) GetValidators(ctx context.Context, pubkeys []string) (map[string]*Validator, error) {
	response, err := c.getValidators(ctx, pubkeys)
	if err != nil {
		return nil, err
	}

	if response.Status != "OK" {
		return nil, fmt.Errorf("error response from server: %s", response.Status)
	}

	validators := make(map[string]*Validator)

	if response.Data == nil {
		return validators, nil
	}

	for i := range response.Data {
		validator := &response.Data[i]
		validators[validator.Pubkey] = validator
	}

	return validators, nil
}

func (c *client) GetValidator(ctx context.Context, pubkey string) (*Validator, error) {
	response, err := c.getValidator(ctx, pubkey)
	if err != nil {
		return nil, err
	}

	if response.Status != "OK" {
		return nil, fmt.Errorf("error response from server: %s", response.Status)
	}

	return &response.Data, nil
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
