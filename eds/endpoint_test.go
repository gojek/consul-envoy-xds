package eds_test

import (
	"testing"

	"github.com/gojektech/consul-envoy-xds/eds"
	"github.com/gojektech/consul-envoy-xds/pubsub"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	cpcore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockConsulAgent struct {
	mock.Mock
}

func (m MockConsulAgent) Locality() *cpcore.Locality {
	args := m.Called()
	return args.Get(0).(*cpcore.Locality)
}

func (m MockConsulAgent) CatalogServiceEndpoints(serviceName string) ([]*api.CatalogService, error) {
	args := m.Called(serviceName)
	return args.Get(0).([]*api.CatalogService), args.Error(1)
}

func (m MockConsulAgent) WatchParams() map[string]string {
	args := m.Called()
	return args.Get(0).(map[string]string)
}

func TestShouldHaveClusterUsingAgentCatalogServiceEndpoints(t *testing.T) {
	agent := &MockConsulAgent{}
	agent.On("CatalogServiceEndpoints", "foo-service").Return([]*api.CatalogService{
		{ServiceName: "foo-service", ServiceAddress: "foo1", ServicePort: 1234},
		{ServiceName: "foo-service", ServiceAddress: "foo2", ServicePort: 1234}},
		nil)
	endpoint := eds.NewEndpoint("foo-service", agent)
	clusters := endpoint.Clusters()

	assert.Equal(t, "foo-service", clusters[0].Name)
	assert.Equal(t, cp.Cluster_USE_DOWNSTREAM_PROTOCOL, clusters[0].ProtocolSelection)
}

func TestShouldHaveCLAUsingAgentCatalogServiceEndpoints(t *testing.T) {
	agent := &MockConsulAgent{}
	agent.On("Locality").Return(&cpcore.Locality{Region: "foo-region"})
	agent.On("CatalogServiceEndpoints", "foo-service").Return([]*api.CatalogService{{ServiceAddress: "foo1", ServicePort: 1234}, {ServiceAddress: "foo2", ServicePort: 1234}}, nil)
	endpoint := eds.NewEndpoint("foo-service", agent)
	cla := endpoint.CLA()

	assert.Equal(t, "foo-service", cla.ClusterName)
	assert.Equal(t, float64(0), cla.Policy.DropOverload)
	localityEndpoint := cla.Endpoints[0]
	assert.Equal(t, "foo-region", localityEndpoint.Locality.Region)
	assert.Equal(t, "socket_address:<address:\"foo1\" port_value:1234 > ", localityEndpoint.LbEndpoints[0].Endpoint.Address.String())
	assert.Equal(t, "socket_address:<address:\"foo2\" port_value:1234 > ", localityEndpoint.LbEndpoints[1].Endpoint.Address.String())
}

func TestShouldSetAgentBasedWatcherParamsInEndpointWatchPlan(t *testing.T) {
	agent := &MockConsulAgent{}
	endpoint := eds.NewEndpoint("foo-service", agent)
	agent.On("WatchParams").Return(map[string]string{"datacenter": "dc-foo-01", "token": "token-foo-01"})

	plan, _ := endpoint.WatchPlan(func(*pubsub.Event) {
	})

	assert.Equal(t, "dc-foo-01", plan.Datacenter)
	assert.Equal(t, "token-foo-01", plan.Token)
}

func TestShouldSetEndpointWatchPlanHandler(t *testing.T) {
	agent := &MockConsulAgent{}
	agent.On("Locality").Return(&cpcore.Locality{Region: "foo-region"})
	agent.On("CatalogServiceEndpoints", "foo-service").Return([]*api.CatalogService{{ServiceAddress: "foo1", ServicePort: 1234}, {ServiceAddress: "foo2", ServicePort: 1234}}, nil)

	endpoint := eds.NewEndpoint("foo-service", agent)
	agent.On("WatchParams").Return(map[string]string{"datacenter": "dc-foo-01", "token": "token-foo-01"})
	var handlerCalled bool
	var capture *cp.ClusterLoadAssignment
	plan, _ := endpoint.WatchPlan(func(event *pubsub.Event) {
		handlerCalled = true
		capture = event.CLA
	})

	plan.Handler(112345, nil)
	assert.True(t, handlerCalled)
	assert.Equal(t, "foo-service", capture.ClusterName)
}
