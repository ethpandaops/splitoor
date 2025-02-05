package notifier

import (
	"context"
	"time"

	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier/source"
	"github.com/sirupsen/logrus"
)

type Publisher struct {
	log     logrus.FieldLogger
	sources []SourceWithConfig
}

type SourceWithConfig struct {
	source source.Source
	group  *string
}

func NewPublisher(ctx context.Context, log logrus.FieldLogger, monitor string, conf Config) (*Publisher, error) {
	sources, err := createSources(ctx, log, monitor, conf.Docs, conf.Sources)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		log:     log,
		sources: sources,
	}, nil
}

func createSources(ctx context.Context, log logrus.FieldLogger, monitor string, docs *string, conf []source.Config) ([]SourceWithConfig, error) {
	sources := make([]SourceWithConfig, len(conf))

	for i, src := range conf {
		s, err := source.NewSource(ctx, log, monitor, src.Name, docs, src.SourceType, src.IncludeMonitorName, src.Group == nil, src.Config)
		if err != nil {
			return nil, err
		}

		sources[i] = SourceWithConfig{
			source: s,
			group:  src.Group,
		}
	}

	return sources, nil
}

func (p *Publisher) Publish(e event.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	for _, src := range p.sources {
		if src.group != nil && e.GetGroup() != *src.group {
			continue
		}

		if err := src.source.Publish(ctx, e); err != nil {
			return err
		}
	}

	return nil
}

func (p *Publisher) Start(ctx context.Context) error {
	for _, src := range p.sources {
		if err := src.source.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (p *Publisher) Stop(ctx context.Context) error {
	for _, src := range p.sources {
		if err := src.source.Stop(ctx); err != nil {
			return err
		}
	}

	return nil
}
