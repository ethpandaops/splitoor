package service

import (
	"context"

	"github.com/ethpandaops/splitoor/pkg/ethereum"
	"github.com/ethpandaops/splitoor/pkg/monitor/beaconchain"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier"
	"github.com/ethpandaops/splitoor/pkg/monitor/safe"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/validator"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Type string

const (
	ServiceTypeUnknown   Type = "unknown"
	ServiceTypeSplit     Type = split.ServiceType
	ServiceTypeValidator Type = validator.ServiceType
)

func CreateServices(ctx context.Context, log logrus.FieldLogger, monitor string, cfg *Config, ethereumPool *ethereum.Pool, publisher *notifier.Publisher, beaconchainClient beaconchain.Client, safeClient safe.Client) ([]Service, error) {
	services := []Service{}

	if cfg.Split != nil {
		sp, err := split.NewService(ctx, log, monitor, cfg.Split, ethereumPool, publisher, safeClient)
		if err != nil {
			return nil, err
		}

		services = append(services, sp)
	}

	if cfg.Validator != nil {
		vp, err := validator.NewService(ctx, log, monitor, cfg.Validator, ethereumPool, publisher, beaconchainClient)
		if err != nil {
			return nil, err
		}

		services = append(services, vp)
	}

	return services, nil
}
