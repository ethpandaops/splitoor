package notifier

import (
	"context"

	"github.com/ethpandaops/splitoor/pkg/monitor/notifier/source"
	"github.com/sirupsen/logrus"
)

type Publisher struct {
	log     logrus.FieldLogger
	sources []source.Source
}

func NewPublisher(ctx context.Context, log logrus.FieldLogger, monitor string, conf Config) (*Publisher, error) {
	sources, err := createSources(ctx, log, monitor, conf.Sources)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		log:     log,
		sources: sources,
	}, nil
}

func createSources(ctx context.Context, log logrus.FieldLogger, monitor string, conf []source.Config) ([]source.Source, error) {
	sources := make([]source.Source, len(conf))

	for i, src := range conf {
		s, err := source.NewSource(ctx, log, monitor, src.Name, src.SourceType, src.Config)
		if err != nil {
			return nil, err
		}

		sources[i] = s
	}

	return sources, nil
}

func (p *Publisher) Publish(ctx context.Context, msg string) error {
	for _, source := range p.sources {
		if err := source.Publish(ctx, msg); err != nil {
			return err
		}
	}

	return nil
}

func (p *Publisher) Start(ctx context.Context) error {
	for _, source := range p.sources {
		if err := source.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (p *Publisher) Stop(ctx context.Context) error {
	for _, source := range p.sources {
		if err := source.Stop(ctx); err != nil {
			return err
		}
	}

	return nil
}
