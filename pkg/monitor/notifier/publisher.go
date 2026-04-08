package notifier

import (
	"context"
	"errors"
	"fmt"
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
	// Create a new background context for this specific call
	return p.PublishWithContext(context.Background(), e)
}

func (p *Publisher) PublishWithContext(ctx context.Context, e event.Event) error {
	if e == nil {
		return errors.New("cannot publish nil event")
	}

	if ctx == nil {
		return errors.New("cannot publish with nil context")
	}

	// Create a reasonable timeout for publishing events
	// This ensures the operation doesn't hang indefinitely, but respects parent cancellation
	publishCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel() // Ensure resources are cleaned up

	for _, src := range p.sources {
		// Check if the context was cancelled before proceeding
		select {
		case <-publishCtx.Done():
			return fmt.Errorf("publishing context cancelled before completion: %w", publishCtx.Err())
		default:
		}

		// Skip nil sources
		if src.source == nil {
			p.log.Warn("Skipping nil source in publisher")

			continue
		}

		// Check if source has a group filter and the event has a group
		eventGroup := e.GetGroup()
		if src.group != nil && eventGroup != *src.group {
			continue
		}

		// Use the timeout context for the publish operation
		if err := src.source.Publish(publishCtx, e); err != nil {
			return fmt.Errorf("error publishing to source %v: %w", src.source.GetName(), err)
		}
	}

	return nil
}

func (p *Publisher) Start(ctx context.Context) error {
	for _, src := range p.sources {
		// Skip nil sources
		if src.source == nil {
			p.log.Warn("Skipping nil source in publisher during start")

			continue
		}

		if err := src.source.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (p *Publisher) Stop(ctx context.Context) error {
	for _, src := range p.sources {
		// Skip nil sources
		if src.source == nil {
			p.log.Warn("Skipping nil source in publisher during stop")

			continue
		}

		if err := src.source.Stop(ctx); err != nil {
			return err
		}
	}

	return nil
}
