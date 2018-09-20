package stream

import (
	cp "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/gojektech/consul-envoy-xds/eds"
	"github.com/gojektech/consul-envoy-xds/pubsub"
)

//ConsulEDS is an implementation of envoy EDS grpc api via envoy go control plan api contract.
type ConsulEDS struct {
	hub     pubsub.Hub
	service eds.Endpoint
}

//StreamAggregatedResources is a grpc streaming api for streaming Discovery responses
func (e *ConsulEDS) StreamAggregatedResources(s cp.AggregatedDiscoveryService_StreamAggregatedResourcesServer) error {
	subscription := e.hub.Subscribe()
	return NewSubscriptionStream(s, subscription, e.service, e.hub).Stream()
}

func New(hub pubsub.Hub, service eds.Endpoint) *ConsulEDS {
	return &ConsulEDS{hub: hub, service: service}
}
