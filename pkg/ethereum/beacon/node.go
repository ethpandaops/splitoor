package beacon

import (
	"context"
	"fmt"
	"sync"
	"time"

	bn "github.com/ethpandaops/beacon/pkg/beacon"
	"github.com/ethpandaops/splitoor/pkg/ethereum/beacon/services"
	"github.com/go-co-op/gocron"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Node struct {
	config *Config
	log    logrus.FieldLogger

	beacon bn.Node

	services []services.Service

	onReadyCallbacks []func(ctx context.Context) error
}

func NewNode(ctx context.Context, log logrus.FieldLogger, name string, config *Config) *Node {
	opts := *bn.
		DefaultOptions()

	opts.HealthCheck.Interval.Duration = time.Second * 3
	opts.HealthCheck.SuccessfulResponses = 1
	opts.PrometheusMetrics = true
	opts.BeaconSubscription.Disable()

	log = log.WithFields(logrus.Fields{"module": "ethereum/beacon", "name": name, "source": config.Name})

	node := bn.NewNode(log, &bn.Config{
		Name:    name,
		Addr:    config.NodeAddress,
		Headers: config.NodeHeaders,
	}, "splitoor_ethereum_beacon", opts)

	metadata := services.NewMetadataService(log, node)

	svcs := []services.Service{
		&metadata,
	}

	return &Node{
		config:   config,
		log:      log,
		beacon:   node,
		services: svcs,
	}
}

func (b *Node) Start(ctx context.Context) error {
	s := gocron.NewScheduler(time.Local)

	// Create a new context with cancellation to manage goroutine lifecycle
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel() // Ensure resources are cleaned up on function exit

	errs := make(chan error, 1)
	done := make(chan struct{})

	go func() {
		defer close(done) // Signal the goroutine is finished when exiting

		wg := sync.WaitGroup{}

		for _, service := range b.services {
			// Check if context is cancelled before starting next service
			select {
			case <-ctxWithCancel.Done():
				return
			default:
			}

			wg.Add(1)

			service.OnReady(ctxWithCancel, func(ctx context.Context) error {
				b.log.WithField("service", service.Name()).Info("Service is ready")

				wg.Done()

				return nil
			})

			b.log.WithField("service", service.Name()).Info("Starting service")

			if err := service.Start(ctxWithCancel); err != nil {
				errs <- fmt.Errorf("failed to start service: %w", err)

				return
			}

			wg.Wait()
		}

		b.log.Info("All services are ready")

		for _, callback := range b.onReadyCallbacks {
			// Check for context cancellation between callbacks
			select {
			case <-ctxWithCancel.Done():
				return
			default:
			}

			if err := callback(ctxWithCancel); err != nil {
				errs <- fmt.Errorf("failed to run on ready callback: %w", err)

				return
			}
		}
	}()

	// Use the context for scheduler as well
	s.StartAsync()

	// Make sure to clean up scheduler
	defer func() {
		s.Stop()
	}()

	if err := b.beacon.Start(ctx); err != nil {
		return err
	}

	// Wait for either an error, completion, or context cancellation
	select {
	case err := <-errs:
		return err
	case <-done:
		// All tasks completed normally
		return nil
	case <-ctx.Done():
		b.log.Info("Context cancelled, stopping beacon node")

		return ctx.Err()
	}
}

func (b *Node) Node() bn.Node {
	return b.beacon
}

func (b *Node) Name() string {
	return b.config.Name
}

func (b *Node) getServiceByName(name services.Name) (services.Service, error) {
	for _, service := range b.services {
		if service.Name() == name {
			return service, nil
		}
	}

	return nil, errors.New("service not found")
}

func (b *Node) Metadata() *services.MetadataService {
	service, err := b.getServiceByName("metadata")
	if err != nil {
		// This should never happen. If it does, good luck.
		return nil
	}

	// Safe type assertion with check
	metadataService, ok := service.(*services.MetadataService)
	if !ok {
		b.log.WithField("service", service).Error("failed to cast service to MetadataService")

		return nil
	}

	return metadataService
}

func (b *Node) OnReady(_ context.Context, callback func(ctx context.Context) error) {
	b.onReadyCallbacks = append(b.onReadyCallbacks, callback)
}

func (b *Node) Synced(ctx context.Context) error {
	// Check if beacon client is available
	if b.beacon == nil {
		return errors.New("beacon client is not initialized")
	}

	status := b.beacon.Status()
	if status == nil {
		return errors.New("missing beacon status")
	}

	syncState := status.SyncState()
	if syncState == nil {
		return errors.New("missing beacon node status sync state")
	}

	if syncState.SyncDistance > 3 {
		return errors.New("beacon node is not synced")
	}

	// Get metadata with nil check
	metadata := b.Metadata()
	if metadata == nil {
		return errors.New("missing metadata service")
	}

	wallclock := metadata.Wallclock()
	if wallclock == nil {
		return errors.New("missing wallclock")
	}

	currentSlot := wallclock.Slots().Current()

	if currentSlot.Number()-uint64(syncState.HeadSlot) > 64 {
		return fmt.Errorf("beacon node is too far behind head, head slot is %d, current slot is %d", syncState.HeadSlot, currentSlot.Number())
	}

	for _, service := range b.services {
		if err := service.Ready(ctx); err != nil {
			return errors.Wrapf(err, "service %s is not ready", service.Name())
		}
	}

	return nil
}
