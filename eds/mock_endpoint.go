package eds

import (
	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/gojek/consul-envoy-xds/pubsub"
	"github.com/hashicorp/consul/watch"
	"github.com/stretchr/testify/mock"
)

type MockEndpoint struct {
	mock.Mock
}

func (m *MockEndpoint) CLA() []*cp.ClusterLoadAssignment {
	args := m.Called()
	return args.Get(0).([]*cp.ClusterLoadAssignment)
}

func (m *MockEndpoint) Clusters() []*cp.Cluster {
	args := m.Called()
	return args.Get(0).([]*cp.Cluster)
}

func (m *MockEndpoint) Routes() []*cp.RouteConfiguration {
	args := m.Called()
	return args.Get(0).([]*cp.RouteConfiguration)
}

func (m *MockEndpoint) WatchPlan(publish func(*pubsub.Event)) (*watch.Plan, error) {
	return nil, nil
}
