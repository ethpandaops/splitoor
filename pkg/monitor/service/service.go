package service

import (
	"context"

	"github.com/ethpandaops/splitoor/pkg/ethereum"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier"
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

func CreateServices(ctx context.Context, log logrus.FieldLogger, monitor string, cfg *Config, ethereumPool *ethereum.Pool, publisher *notifier.Publisher) ([]Service, error) {
	services := []Service{}

	sp, err := split.NewService(ctx, log, monitor, &cfg.Split, ethereumPool, publisher)
	if err != nil {
		return nil, err
	}

	services = append(services, sp)

	vp, err := validator.NewService(ctx, log, monitor, &cfg.Validator, ethereumPool, publisher)
	if err != nil {
		return nil, err
	}

	services = append(services, vp)

	return services, nil
}
