package source

import (
	"context"
	"errors"

	"github.com/creasty/defaults"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier/source/discord"

	"github.com/sirupsen/logrus"
)

type Source interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Type() string
	Name() string
	Publish(ctx context.Context, msg string) error
}

func NewSource(ctx context.Context, log logrus.FieldLogger, monitor, sourceName string, sourceType SourceType, config *RawMessage) (Source, error) {
	if sourceType == SourceTypeUnknown {
		return nil, errors.New("source type is required")
	}

	switch sourceType {
	case SourceTypeDiscord:
		conf := &discord.Config{}

		if config != nil {
			if err := config.Unmarshal(conf); err != nil {
				return nil, err
			}
		}

		if err := defaults.Set(conf); err != nil {
			return nil, err
		}

		return discord.NewDiscord(ctx, log, monitor, sourceName, conf)
	}

	return nil, errors.New("source type is not supported")
}
