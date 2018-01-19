package eds

import (
	"github.com/gojektech/consul-envoy-xds/pubsub"

	cp "github.com/envoyproxy/go-control-plane/api"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//ConsulEDS is an implementation of envoy EDS grpc api via envoy go control plan api contract.
type ConsulEDS struct {
	hub            pubsub.Hub
	watchedService Endpoint
}

//StreamEndpoints is a grpc streaming api for streaming Discovery responses
func (e *ConsulEDS) StreamEndpoints(stream cp.EndpointDiscoveryService_StreamEndpointsServer) error {
	subscription := e.hub.Subscribe()
	subscription.Accept(e.watchedService.CLA())
	return NewSubscriptionStream(stream, subscription).Stream()
}

func (e *ConsulEDS) StreamLoadStats(stream cp.EndpointDiscoveryService_StreamLoadStatsServer) error {
	return nil
}

func (e *ConsulEDS) FetchEndpoints(context.Context, *cp.DiscoveryRequest) (*cp.DiscoveryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func New(hub pubsub.Hub, svc Endpoint) *ConsulEDS {
	return &ConsulEDS{hub: hub, watchedService: svc}
}
