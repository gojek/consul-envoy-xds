package edswatch

import (
	"github.com/gojektech/consul-envoy-xds/eds"
	"github.com/gojektech/consul-envoy-xds/pubsub"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/hashicorp/consul/watch"
)

// ServiceWatch runs consul watch for the specified service
type ServiceWatch struct {
	service   eds.Endpoint
	plan      *watch.Plan
	agentHost string
	hub       pubsub.Hub
}

//PublishCLA publishes a cluster load assignment to hub
func (sw *ServiceWatch) PublishCLA(idx uint64, data interface{}) {
	sw.hub.Publish(&pubsub.Event{sw.service.CLA(), []*cp.Cluster{&cp.Cluster{}}, []*cp.RouteConfiguration{&cp.RouteConfiguration{}}})
}

// NewWatch creates a new service watch
func NewWatch(agentHost string, service eds.Endpoint, hub pubsub.Hub) (*ServiceWatch, error) {
	watchPlan, err := service.WatchPlan(hub.Publish)

	if err != nil {
		return nil, err
	}

	return &ServiceWatch{service: service, plan: watchPlan, agentHost: agentHost, hub: hub}, nil
}

//Run consul watch for the specified service
func (sw ServiceWatch) Run(errorChannel chan error) {
	//Initial publish
	if err := sw.plan.Run(sw.agentHost); err != nil {
		errorChannel <- err
	}
}
