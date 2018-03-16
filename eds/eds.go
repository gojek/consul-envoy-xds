package eds

import (
	"github.com/gojektech/consul-envoy-xds/pubsub"
	"github.com/gojektech/consul-envoy-xds/stream"

	cp "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
)

//ConsulEDS is an implementation of envoy EDS grpc api via envoy go control plan api contract.
type ConsulEDS struct {
	hub            pubsub.Hub
	watchedService Endpoint
}

//StreamAggregatedResources is a grpc streaming api for streaming Discovery responses
func (e *ConsulEDS) StreamAggregatedResources(s cp.AggregatedDiscoveryService_StreamAggregatedResourcesServer) error {
	subscription := e.hub.Subscribe()
	return stream.NewSubscriptionStream(s, subscription).Stream()
}

func New(hub pubsub.Hub, svc Endpoint) *ConsulEDS {
	return &ConsulEDS{hub: hub, watchedService: svc}
}
