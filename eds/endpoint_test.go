package eds_test

import (
	"testing"

	"github.com/gojektech/consul-envoy-xds/eds"
	"github.com/gojektech/consul-envoy-xds/pubsub"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	cpcore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/gojektech/consul-envoy-xds/config"
)

type MockConsulAgent struct {
	mock.Mock
}

func (m *MockConsulAgent) Locality() *cpcore.Locality {
	args := m.Called()
	return args.Get(0).(*cpcore.Locality)
}

func (m *MockConsulAgent) CatalogServiceEndpoints(serviceName ...string) ([][]*api.CatalogService, error) {
	args := m.Called(serviceName)
	return args.Get(0).([][]*api.CatalogService), args.Error(1)
}

func (m *MockConsulAgent) WatchParams() map[string]string {
	args := m.Called()
	return args.Get(0).(map[string]string)
}

func TestShouldHaveClusterUsingAgentCatalogServiceEndpoints(t *testing.T) {
	agent := &MockConsulAgent{}
	agent.On("CatalogServiceEndpoints", []string{"foo-service"}).Return([][]*api.CatalogService{[]*api.CatalogService{
		{ServiceName: "foo-service", ServiceAddress: "foo1", ServicePort: 1234},
		{ServiceName: "foo-service", ServiceAddress: "foo2", ServicePort: 1234}}},
		nil)
	httpRateLimitConfig := &config.HTTPHeaderRateLimitConfig{IsEnabled: false}
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo-service", Whitelist: []string{"/"}}}, agent, httpRateLimitConfig)
	clusters := endpoint.Clusters()

	assert.Equal(t, "foo-service", clusters[0].Name)
	assert.Equal(t, cp.Cluster_USE_DOWNSTREAM_PROTOCOL, clusters[0].ProtocolSelection)
}

func TestShouldHaveAuthHeaderRateLimit(t *testing.T) {
	agent := &MockConsulAgent{}

	httpRateLimitConfig := &config.HTTPHeaderRateLimitConfig{IsEnabled: true, HeaderName: "authorization", DescriptorKey: "auth_token"}
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo", Whitelist: []string{"/hoo", "/bar"}}}, agent, httpRateLimitConfig)

	routeConfig := endpoint.Routes()
	virtualHosts := routeConfig[0].GetVirtualHosts()
	assert.Equal(t, "local_route", routeConfig[0].GetName())
	assert.Equal(t, []string{"*"}, virtualHosts[0].GetDomains())

	actionSpecifier := route.RateLimit_Action_RequestHeaders_{
		RequestHeaders: &route.RateLimit_Action_RequestHeaders{
			HeaderName:    "authorization",
			DescriptorKey: "auth_token",
		},
	}

	routeAction := route.RateLimit_Action{
		ActionSpecifier: &actionSpecifier,
	}

	rateLimit := &route.RateLimit{
		Actions: []*route.RateLimit_Action{&routeAction},
	}

	var rateLimits []*route.RateLimit
	rateLimits = append(rateLimits, rateLimit)

	assert.ElementsMatch(t, rateLimits, virtualHosts[0].RateLimits)
}

func TestMultipleRouteConfiguration(t *testing.T) {
	agent := &MockConsulAgent{}

	httpRateLimitConfig := &config.HTTPHeaderRateLimitConfig{IsEnabled: false}
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo", Whitelist: []string{"/hoo", "/bar"}}}, agent, httpRateLimitConfig)

	routeConfig := endpoint.Routes()
	virtualHosts := routeConfig[0].GetVirtualHosts()
	assert.Equal(t, "local_route", routeConfig[0].GetName())
	assert.Equal(t, []string{"*"}, virtualHosts[0].GetDomains())
	assert.Equal(t, 1, len(virtualHosts))
	assert.Equal(t, 2, len(virtualHosts[0].GetRoutes()))
	assert.ElementsMatch(t, []route.Route{{
		Match: route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: "/hoo",
			},
		},
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: "foo",
				},
			},
		},
	}, {
		Match: route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: "/bar",
			},
		},
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: "foo",
				},
			},
		},
	}}, virtualHosts[0].GetRoutes())
}

func TestShouldHaveCLAUsingAgentCatalogServiceEndpoints(t *testing.T) {
	agent := &MockConsulAgent{}
	agent.On("Locality").Return(&cpcore.Locality{Region: "foo-region"})
	agent.On("CatalogServiceEndpoints", []string{"foo-service"}).Return([][]*api.CatalogService{
		[]*api.CatalogService{{ServiceName: "foo-service", ServiceAddress: "foo1", ServicePort: 1234}, {ServiceAddress: "foo2", ServicePort: 1234}}},
		nil,
	)
	httpRateLimitConfig := &config.HTTPHeaderRateLimitConfig{IsEnabled: false}
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo-service", Whitelist: []string{"/"}}}, agent, httpRateLimitConfig)
	cla := endpoint.CLA()[0]

	assert.Equal(t, "foo-service", cla.ClusterName)
	assert.Equal(t, float64(0), cla.Policy.DropOverload)
	localityEndpoint := cla.Endpoints[0]
	assert.Equal(t, "foo-region", localityEndpoint.Locality.Region)
	assert.Equal(t, "socket_address:<address:\"foo1\" port_value:1234 > ", localityEndpoint.LbEndpoints[0].Endpoint.Address.String())
	assert.Equal(t, "socket_address:<address:\"foo2\" port_value:1234 > ", localityEndpoint.LbEndpoints[1].Endpoint.Address.String())
}

func TestShouldSetAgentBasedWatcherParamsInEndpointWatchPlan(t *testing.T) {
	agent := &MockConsulAgent{}
	httpRateLimitConfig := &config.HTTPHeaderRateLimitConfig{IsEnabled: false}
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo-service", Whitelist: []string{"/"}}}, agent, httpRateLimitConfig)
	agent.On("WatchParams").Return(map[string]string{"datacenter": "dc-foo-01", "token": "token-foo-01"})

	plan, _ := endpoint.WatchPlan(func(*pubsub.Event) {
	})

	assert.Equal(t, "dc-foo-01", plan.Datacenter)
	assert.Equal(t, "token-foo-01", plan.Token)
	assert.Equal(t, "services", plan.Type)
}

func TestShouldSetEndpointWatchPlanHandler(t *testing.T) {
	agent := &MockConsulAgent{}
	agent.On("Locality").Return(&cpcore.Locality{Region: "foo-region"})
	agent.On("CatalogServiceEndpoints", []string{"foo-service"}).Return([][]*api.CatalogService{
		[]*api.CatalogService{{ServiceName: "foo-service", ServiceAddress: "foo1", ServicePort: 1234}, {ServiceAddress: "foo2", ServicePort: 1234}},
	}, nil)

	httpRateLimitConfig := &config.HTTPHeaderRateLimitConfig{IsEnabled: false}
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo-service", Whitelist: []string{"/"}}}, agent, httpRateLimitConfig)
	agent.On("WatchParams").Return(map[string]string{"datacenter": "dc-foo-01", "token": "token-foo-01"})
	var handlerCalled bool
	var capture *cp.ClusterLoadAssignment
	plan, _ := endpoint.WatchPlan(func(event *pubsub.Event) {
		handlerCalled = true
		capture = event.CLA[0]
	})

	plan.Handler(112345, nil)
	assert.True(t, handlerCalled)
	assert.Equal(t, "foo-service", capture.ClusterName)
}
