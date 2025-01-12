package execution

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution/services"
	"github.com/sirupsen/logrus"
)

// headerTransport adds custom headers to requests
type headerTransport struct {
	headers map[string]string
	base    http.RoundTripper
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range t.headers {
		req.Header.Set(key, value)
	}

	return t.base.RoundTrip(req)
}

type Node struct {
	config *Config
	log    logrus.FieldLogger
	rpc    *ethrpc.Provider
	name   string

	services []services.Service

	onReadyCallbacks []func(ctx context.Context) error
}

func NewNode(log logrus.FieldLogger, name string, conf *Config) *Node {
	return &Node{
		config:   conf,
		name:     name,
		log:      log.WithFields(logrus.Fields{"module": "ethereum/beacon", "name": name}),
		services: []services.Service{},
	}
}

func (n *Node) OnReady(_ context.Context, callback func(ctx context.Context) error) {
	n.onReadyCallbacks = append(n.onReadyCallbacks, callback)
}

func (n *Node) Start(ctx context.Context) error {
	httpClient := http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	httpClient.Transport = &headerTransport{
		headers: n.config.NodeHeaders,
		base:    httpClient.Transport,
	}

	rpc, err := ethrpc.NewProvider(n.config.NodeAddress, ethrpc.WithHTTPClient(&httpClient))
	if err != nil {
		return err
	}

	metadata := services.NewMetadataService(n.log, rpc)

	svcs := []services.Service{
		&metadata,
	}

	n.rpc = rpc

	n.services = svcs

	errs := make(chan error, 1)

	go func() {
		wg := sync.WaitGroup{}

		for _, service := range n.services {
			wg.Add(1)

			service.OnReady(ctx, func(ctx context.Context) error {
				n.log.WithField("service", service.Name()).Info("Service is ready")

				wg.Done()

				return nil
			})

			n.log.WithField("service", service.Name()).Info("Starting service")

			if err := service.Start(ctx); err != nil {
				errs <- fmt.Errorf("failed to start service: %w", err)
			}

			wg.Wait()
		}

		n.log.Info("All services are ready")

		for _, callback := range n.onReadyCallbacks {
			if err := callback(ctx); err != nil {
				errs <- fmt.Errorf("failed to run on ready callback: %w", err)
			}
		}
	}()

	return nil
}

func (n *Node) Stop() error {
	return nil
}

func (n *Node) getServiceByName(name services.Name) (services.Service, error) {
	for _, service := range n.services {
		if service.Name() == name {
			return service, nil
		}
	}

	return nil, errors.New("service not found")
}

func (n *Node) Metadata() *services.MetadataService {
	service, err := n.getServiceByName("metadata")
	if err != nil {
		// This should never happen. If it does, good luck.
		return nil
	}

	return service.(*services.MetadataService)
}

func (n *Node) Name() string {
	return n.config.Name
}
