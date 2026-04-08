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

type configValidator interface {
	Validate() error
}

func initConfig[T configValidator](config *RawMessage, conf T) error {
	if config != nil {
		if err := config.Unmarshal(conf); err != nil {
			return err
		}
	}

	if err := defaults.Set(conf); err != nil {
		return err
	}

	return conf.Validate()
}

func NewSource(ctx context.Context, log logrus.FieldLogger, monitor, sourceName string, docs *string, sourceType SourceType, includeMonitorName, includeGroupName bool, config *RawMessage) (Source, error) {
	if sourceType == SourceTypeUnknown {
		return nil, errors.New("source type is required")
	}

	switch sourceType { //nolint:exhaustive // SourceTypeUnknown is handled above
	case SourceTypeDiscord:
		conf := &discord.Config{}
		if err := initConfig(config, conf); err != nil {
			return nil, err
		}

		return discord.NewDiscord(ctx, log, monitor, sourceName, docs, includeMonitorName, includeGroupName, conf)
	case SourceTypeSMTP:
		conf := &smtp.Config{}
		if err := initConfig(config, conf); err != nil {
			return nil, err
		}

		return smtp.NewSMTP(ctx, log, monitor, sourceName, docs, includeMonitorName, includeGroupName, conf)
	case SourceTypeSES:
		conf := &ses.Config{}
		if err := initConfig(config, conf); err != nil {
			return nil, err
		}

		return ses.NewSES(ctx, log, monitor, sourceName, docs, includeMonitorName, includeGroupName, conf)
	case SourceTypeTelegram:
		conf := &telegram.Config{}
		if err := initConfig(config, conf); err != nil {
			return nil, err
		}

		return telegram.NewTelegram(ctx, log, monitor, sourceName, docs, includeMonitorName, includeGroupName, conf)
	}

	return nil, errors.New("source type is not supported")
}
