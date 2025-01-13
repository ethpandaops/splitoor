package source

import (
	"context"
	"errors"

	"github.com/creasty/defaults"
	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier/source/discord"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier/source/ses"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier/source/smtp"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier/source/telegram"

	"github.com/sirupsen/logrus"
)

type Source interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetType() string
	GetName() string
	Publish(ctx context.Context, e event.Event) error
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

		if err := conf.Validate(); err != nil {
			return nil, err
		}

		return discord.NewDiscord(ctx, log, monitor, sourceName, conf)
	case SourceTypeSMTP:
		conf := &smtp.Config{}

		if config != nil {
			if err := config.Unmarshal(conf); err != nil {
				return nil, err
			}
		}

		if err := defaults.Set(conf); err != nil {
			return nil, err
		}

		if err := conf.Validate(); err != nil {
			return nil, err
		}

		return smtp.NewSMTP(ctx, log, monitor, sourceName, conf)
	case SourceTypeSES:
		conf := &ses.Config{}

		if config != nil {
			if err := config.Unmarshal(conf); err != nil {
				return nil, err
			}
		}

		if err := defaults.Set(conf); err != nil {
			return nil, err
		}

		if err := conf.Validate(); err != nil {
			return nil, err
		}

		return ses.NewSES(ctx, log, monitor, sourceName, conf)
	case SourceTypeTelegram:
		conf := &telegram.Config{}

		if config != nil {
			if err := config.Unmarshal(conf); err != nil {
				return nil, err
			}
		}

		if err := defaults.Set(conf); err != nil {
			return nil, err
		}

		if err := conf.Validate(); err != nil {
			return nil, err
		}

		return telegram.NewTelegram(ctx, log, monitor, sourceName, conf)
	}

	return nil, errors.New("source type is not supported")
}
