package controller

import (
	"context"
	"errors"

	"github.com/creasty/defaults"
	"github.com/ethpandaops/splitoor/pkg/ethereum"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier"
	s "github.com/ethpandaops/splitoor/pkg/monitor/safe"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/controller/eoa"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/controller/safe"

	"github.com/sirupsen/logrus"
)

type Controller interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Type() string
	Name() string
	Address() string
}

func NewController(ctx context.Context, log logrus.FieldLogger, monitor, name string, controllerType ControllerType, config *RawMessage, splitAddress, splitsContractAddress string, ethereumPool *ethereum.Pool, safeClient s.Client, publisher *notifier.Publisher) (Controller, error) {
	if controllerType == ControllerTypeUnknown {
		return nil, errors.New("controller type is required")
	}

	switch controllerType {
	case ControllerTypeEOA:
		conf := &eoa.Config{}

		if config != nil {
			if err := config.Unmarshal(conf); err != nil {
				return nil, err
			}
		}

		if err := defaults.Set(conf); err != nil {
			return nil, err
		}

		return eoa.New(ctx, log, name, conf)
	case ControllerTypeSafe:
		conf := &safe.Config{}

		if config != nil {
			if err := config.Unmarshal(conf); err != nil {
				return nil, err
			}
		}

		if err := defaults.Set(conf); err != nil {
			return nil, err
		}

		return safe.New(ctx, log, monitor, name, conf, splitAddress, splitsContractAddress, ethereumPool, safeClient, publisher)
	}

	return nil, errors.New("controller type is not supported")
}
