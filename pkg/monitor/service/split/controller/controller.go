package controller

import (
	"context"
	"errors"

	"github.com/creasty/defaults"
	"github.com/ethpandaops/splitoor/pkg/ethereum"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/controller/eoa"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/controller/safe"

	"github.com/sirupsen/logrus"
)

type Controller interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Type() string
	Name() string
}

func NewController(ctx context.Context, log logrus.FieldLogger, name string, controllerType ControllerType, config *RawMessage, ethereumPool *ethereum.Pool) (Controller, error) {
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

		return safe.New(ctx, log, name, conf, ethereumPool)
	}

	return nil, errors.New("controller type is not supported")
}
