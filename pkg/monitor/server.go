package monitor

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	//nolint:gosec // only exposed if pprofAddr config is set
	_ "net/http/pprof"

	"github.com/ethpandaops/splitoor/pkg/ethereum"
	"github.com/ethpandaops/splitoor/pkg/monitor/beaconchain"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier"
	"github.com/ethpandaops/splitoor/pkg/monitor/safe"
	"github.com/ethpandaops/splitoor/pkg/monitor/service"
	"github.com/ethpandaops/splitoor/pkg/observability"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	log    logrus.FieldLogger
	config *Config

	services  []service.Service
	publisher *notifier.Publisher

	metricsServer *http.Server
	pprofServer   *http.Server
	healthServer  *http.Server

	ethereumPool *ethereum.Pool

	beaconchainClient beaconchain.Client
	safeClient        safe.Client
}

func NewServer(ctx context.Context, log logrus.FieldLogger, conf *Config) (*Server, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	ethereumPool := ethereum.NewPool(ctx, log, conf.Name, &conf.Ethereum)

	publisher, err := notifier.NewPublisher(ctx, log, conf.Name, conf.Notifier)
	if err != nil {
		return nil, err
	}

	var beaconchainClient beaconchain.Client
	if conf.Beaconchain.Enabled {
		beaconchainClient, err = beaconchain.NewClient(ctx, log, conf.Name, &conf.Beaconchain)
		if err != nil {
			return nil, err
		}
	}

	var safeClient safe.Client
	if conf.Safe.Enabled {
		safeClient, err = safe.NewClient(ctx, log, conf.Name, &conf.Safe)
		if err != nil {
			return nil, err
		}
	}

	services, err := service.CreateServices(ctx, log, conf.Name, &conf.Services, ethereumPool, publisher, beaconchainClient, safeClient)
	if err != nil {
		return nil, err
	}

	return &Server{
		config:            conf,
		log:               log.WithField("component", "server"),
		services:          services,
		publisher:         publisher,
		ethereumPool:      ethereumPool,
		beaconchainClient: beaconchainClient,
		safeClient:        safeClient,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	observability.StartMetricsServer(ctx, s.config.MetricsAddr)

	g, ctx := errgroup.WithContext(ctx)

	if s.config.PProfAddr != nil {
		g.Go(func() error {
			if err := s.startPProf(); err != nil {
				if err != http.ErrServerClosed {
					return err
				}
			}

			return nil
		})
	}

	if s.config.HealthCheckAddr != nil {
		g.Go(func() error {
			if err := s.startHealthCheck(); err != nil {
				if err != http.ErrServerClosed {
					return err
				}
			}

			return nil
		})
	}

	g.Go(func() error {
		return s.publisher.Start(ctx)
	})

	g.Go(func() error {
		return s.startServices(ctx)
	})

	g.Go(func() error {
		s.ethereumPool.Start(ctx)

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()

		return s.stop(ctx)
	})

	if err := g.Wait(); err != context.Canceled {
		return err
	}

	return nil
}

func (s *Server) stop(ctx context.Context) error {
	if err := s.publisher.Stop(ctx); err != nil {
		return err
	}

	for _, svc := range s.services {
		if err := svc.Stop(ctx); err != nil {
			return err
		}
	}

	if s.pprofServer != nil {
		if err := s.pprofServer.Shutdown(ctx); err != nil {
			return err
		}
	}

	if s.healthServer != nil {
		if err := s.healthServer.Shutdown(ctx); err != nil {
			return err
		}
	}

	if s.metricsServer != nil {
		if err := s.metricsServer.Shutdown(ctx); err != nil {
			return err
		}
	}

	s.log.Info("Server stopped")

	return nil
}

func (s *Server) startServices(ctx context.Context) error {
	for _, svc := range s.services {
		if err := svc.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) startPProf() error {
	s.log.WithField("addr", *s.config.PProfAddr).Info("Starting pprof server")

	s.pprofServer = &http.Server{
		Addr:              *s.config.PProfAddr,
		ReadHeaderTimeout: 120 * time.Second,
	}

	return s.pprofServer.ListenAndServe()
}

func (s *Server) startHealthCheck() error {
	s.log.WithField("addr", *s.config.HealthCheckAddr).Info("Starting healthcheck server")

	s.healthServer = &http.Server{
		Addr:              *s.config.HealthCheckAddr,
		ReadHeaderTimeout: 120 * time.Second,
	}

	s.healthServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return s.healthServer.ListenAndServe()
}
