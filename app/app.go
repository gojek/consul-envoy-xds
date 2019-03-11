package app

import (
	"fmt"
	"log"
	"net"

	"github.com/gojektech/consul-envoy-xds/config"
	"github.com/gojektech/consul-envoy-xds/eds"
	"github.com/gojektech/consul-envoy-xds/edswatch"
	"github.com/gojektech/consul-envoy-xds/pubsub"

	cp "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/gojektech/consul-envoy-xds/agent"
	"github.com/gojektech/consul-envoy-xds/stream"
	"google.golang.org/grpc"
)

var server *grpc.Server

func Start() {
	cfg := config.Load()
	hub := pubsub.NewHub()
	svcCfg := cfg.WatchedServices()

	var services []eds.Service
	for _, s := range svcCfg {
		services = append(services, eds.Service{
			Name:      s,
			Whitelist: cfg.WhitelistedRoutes(s),
		})
	}

	svc := eds.NewEndpoint(services, agent.NewAgent(cfg.ConsulAddress(), cfg.ConsulToken(), cfg.ConsulDC()), cfg.GetHTTPHeaderRateLimitConfig())
	w, err := edswatch.NewWatch(cfg.ConsulAddress(), svc, hub)
	if err != nil {
		log.Printf("failed to register watch: %v\n", err)
	}
	errChan := make(chan error)
	go w.Run(errChan)
	defer w.Stop()

	server = grpc.NewServer()
	cp.RegisterAggregatedDiscoveryServiceServer(server, stream.New(hub, svc))
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port()))
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}
	server.Serve(lis)
}

func Stop() {
	server.Stop()
}
