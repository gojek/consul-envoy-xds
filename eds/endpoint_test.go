package eds_test

import (
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/cluster"
	"github.com/gojek/consul-envoy-xds/utils"
	"testing"

	"github.com/gojek/consul-envoy-xds/eds"
	"github.com/gojek/consul-envoy-xds/pubsub"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	cpcore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	"github.com/gojek/consul-envoy-xds/config"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (m *MockConsulAgent) HealthCheckCatalogServiceEndpoints(serviceName ...string) ([][]*api.CatalogService, error) {
	args := m.Called(serviceName)
	return args.Get(0).([][]*api.CatalogService), args.Error(1)
}

func (m *MockConsulAgent) WatchParams() map[string]string {
	args := m.Called()
	return args.Get(0).(map[string]string)
}

func TestShouldHaveClusterUsingAgentCatalogServiceEndpoints(t *testing.T) {
	agent := &MockConsulAgent{}
	agent.On("CatalogServiceEndpoints", []string{"foo-service"}).Return([][]*api.CatalogService{{
		{ServiceName: "foo-service", ServiceAddress: "foo1", ServicePort: 1234},
		{ServiceName: "foo-service", ServiceAddress: "foo2", ServicePort: 1234}}},
		nil)

	httpRateLimitConfig := &config.HTTPHeaderRateLimitConfig{IsEnabled: false}
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo-service", Whitelist: []string{"/"}}}, agent, httpRateLimitConfig, false)
	clusters := endpoint.Clusters()

	assert.Equal(t, "foo-service", clusters[0].Name)
	assert.Equal(t, cp.Cluster_USE_DOWNSTREAM_PROTOCOL, clusters[0].ProtocolSelection)
	circuitBreakers := cluster.CircuitBreakers{Thresholds: []*cluster.CircuitBreakers_Thresholds{{
		Priority:           cpcore.RoutingPriority_DEFAULT,
		MaxConnections:     utils.Uint32Value(1024),
		MaxPendingRequests: utils.Uint32Value(1024),
		MaxRequests:        utils.Uint32Value(1024),
		MaxRetries:         utils.Uint32Value(3),
	}}}
	assert.Equal(t, circuitBreakers, clusters[0].CircuitBreakers)
}

func TestShouldHaveAuthHeaderRateLimit(t *testing.T) {
	agent := &MockConsulAgent{}

	httpRateLimitConfig := &config.HTTPHeaderRateLimitConfig{IsEnabled: true, HeaderName: "authorization", DescriptorKey: "auth_token"}
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo", Whitelist: []string{"/hoo", "/bar"}}}, agent, httpRateLimitConfig, false)

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
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo", Whitelist: []string{"/hoo", "/bar"}}}, agent, httpRateLimitConfig, false)

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

func TestMultipleRouteConfigurationWithRegexPath(t *testing.T) {
	agent := &MockConsulAgent{}
	agent.On("CatalogServiceEndpoints", []string{"foo"}).Return([][]*api.CatalogService{{
		{ServiceName: "foo", ServiceAddress: "foo1", ServicePort: 1234}}},
		nil)

	httpRateLimitConfig := &config.HTTPHeaderRateLimitConfig{IsEnabled: false}
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo", Whitelist: []string{"%regex:/hoo", "/bar"}}}, agent, httpRateLimitConfig, false)

	routeConfig := endpoint.Routes()
	virtualHosts := routeConfig[0].GetVirtualHosts()
	assert.Equal(t, "local_route", routeConfig[0].GetName())
	assert.Equal(t, []string{"*"}, virtualHosts[0].GetDomains())
	assert.ElementsMatch(t, []route.Route{{
		Match: route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Regex{
				Regex: "/hoo",
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
		{{ServiceName: "foo-service", ServiceAddress: "foo1", ServicePort: 1234}, {ServiceAddress: "foo2", ServicePort: 1234}}},
		nil,
	)
	httpRateLimitConfig := &config.HTTPHeaderRateLimitConfig{IsEnabled: false}
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo-service", Whitelist: []string{"/"}}}, agent, httpRateLimitConfig, false)
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
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo-service", Whitelist: []string{"/"}}}, agent, httpRateLimitConfig, false)
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
		{{ServiceName: "foo-service", ServiceAddress: "foo1", ServicePort: 1234}, {ServiceAddress: "foo2", ServicePort: 1234}},
	}, nil)

	httpRateLimitConfig := &config.HTTPHeaderRateLimitConfig{IsEnabled: false}
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo-service", Whitelist: []string{"/"}}}, agent, httpRateLimitConfig, false)
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

func TestShouldHaveClusterUsingAgentCatalogServiceEndpointsWhenHealthCheckEnabled(t *testing.T) {
	agent := &MockConsulAgent{}
	agent.On("HealthCheckCatalogServiceEndpoints", []string{"foo-service"}).Return([][]*api.CatalogService{{
		{ServiceName: "foo-service", ServiceAddress: "foo1", ServicePort: 1234},
		{ServiceName: "foo-service", ServiceAddress: "foo2", ServicePort: 1234}}},
		nil)

	httpRateLimitConfig := &config.HTTPHeaderRateLimitConfig{IsEnabled: false}
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo-service", Whitelist: []string{"/"}}}, agent, httpRateLimitConfig, true)
	clusters := endpoint.Clusters()

	assert.Equal(t, "foo-service", clusters[0].Name)
	assert.Equal(t, cp.Cluster_USE_DOWNSTREAM_PROTOCOL, clusters[0].ProtocolSelection)
}

func TestMultipleRouteConfigurationWithRegexPathWhenHealthCheckEnabled(t *testing.T) {
	agent := &MockConsulAgent{}
	agent.On("HealthCheckCatalogServiceEndpoints", []string{"foo"}).Return([][]*api.CatalogService{{
		{ServiceName: "foo", ServiceAddress: "foo1", ServicePort: 1234}}},
		nil)

	httpRateLimitConfig := &config.HTTPHeaderRateLimitConfig{IsEnabled: false}
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo", Whitelist: []string{"%regex:/hoo", "/bar"}}}, agent, httpRateLimitConfig, true)

	routeConfig := endpoint.Routes()
	virtualHosts := routeConfig[0].GetVirtualHosts()
	assert.Equal(t, "local_route", routeConfig[0].GetName())
	assert.Equal(t, []string{"*"}, virtualHosts[0].GetDomains())
	assert.ElementsMatch(t, []route.Route{{
		Match: route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Regex{
				Regex: "/hoo",
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

func TestShouldHaveCLAUsingAgentCatalogServiceEndpointsWhenHealthCheckEnabled(t *testing.T) {
	agent := &MockConsulAgent{}
	agent.On("Locality").Return(&cpcore.Locality{Region: "foo-region"})
	agent.On("HealthCheckCatalogServiceEndpoints", []string{"foo-service"}).Return([][]*api.CatalogService{
		{{ServiceName: "foo-service", ServiceAddress: "foo1", ServicePort: 1234}, {ServiceAddress: "foo2", ServicePort: 1234}}},
		nil,
	)
	httpRateLimitConfig := &config.HTTPHeaderRateLimitConfig{IsEnabled: false}
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo-service", Whitelist: []string{"/"}}}, agent, httpRateLimitConfig, true)
	cla := endpoint.CLA()[0]

	assert.Equal(t, "foo-service", cla.ClusterName)
	assert.Equal(t, float64(0), cla.Policy.DropOverload)
	localityEndpoint := cla.Endpoints[0]
	assert.Equal(t, "foo-region", localityEndpoint.Locality.Region)
	assert.Equal(t, "socket_address:<address:\"foo1\" port_value:1234 > ", localityEndpoint.LbEndpoints[0].Endpoint.Address.String())
	assert.Equal(t, "socket_address:<address:\"foo2\" port_value:1234 > ", localityEndpoint.LbEndpoints[1].Endpoint.Address.String())
}

func TestShouldSetEndpointWatchPlanHandlerWhenHealthCheckEnabled(t *testing.T) {
	agent := &MockConsulAgent{}
	agent.On("Locality").Return(&cpcore.Locality{Region: "foo-region"})
	agent.On("HealthCheckCatalogServiceEndpoints", []string{"foo-service"}).Return([][]*api.CatalogService{
		{{ServiceName: "foo-service", ServiceAddress: "foo1", ServicePort: 1234}, {ServiceAddress: "foo2", ServicePort: 1234}},
	}, nil)

	httpRateLimitConfig := &config.HTTPHeaderRateLimitConfig{IsEnabled: false}
	endpoint := eds.NewEndpoint([]eds.Service{{Name: "foo-service", Whitelist: []string{"/"}}}, agent, httpRateLimitConfig, true)
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
