package main

import (
	"fmt"
	"log"
	"net"

	"github.com/gojektech/consul-envoy-xds/config"
	"github.com/gojektech/consul-envoy-xds/eds"
	"github.com/gojektech/consul-envoy-xds/edswatch"
	"github.com/gojektech/consul-envoy-xds/pubsub"

	cp "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"google.golang.org/grpc"
)

func main() {

	cfg := config.Load()
	hub := pubsub.NewHub()
	svc := cfg.WatchedService()
	w, _ := edswatch.NewWatch(cfg.ConsulAddress(), svc, hub)
	errChan := make(chan error)
	go w.Run(errChan)

	g := grpc.NewServer()
	cp.RegisterAggregatedDiscoveryServiceServer(g, eds.New(hub, svc))
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port()))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	g.Serve(lis)
}
