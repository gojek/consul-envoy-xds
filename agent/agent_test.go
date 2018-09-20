package agent_test

import (
	"testing"

	"github.com/gojektech/consul-envoy-xds/agent"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

func TestShouldUseDatacenterAsRegionForTheControlPlaneEndpoint(t *testing.T) {
	a := agent.NewAgent("localhost:8500", "", "dc1")
	locality := a.Locality()
	assert.Equal(t, "dc1", locality.Region)
}

func TestShouldHaveDCAndTokenInWatcherParams(t *testing.T) {
	a := agent.NewAgent("localhost:8500", "Foo", "dc1")
	assert.Equal(t, map[string]string{"datacenter": "dc1", "token": "Foo"}, a.WatchParams())
}

func TestShouldReturnServicesFromCatalog(t *testing.T) {
	consulSvr, consulClient := agent.StartConsulTestServer()
	defer consulSvr.Stop()

	consulClient.Agent().ServiceRegister(&api.AgentServiceRegistration{Name: "foo", ID: "foo"})
	consulClient.Agent().ServiceRegister(&api.AgentServiceRegistration{Name: "bar", ID: "bar"})

	a := agent.NewAgent(consulSvr.HTTPAddr, "Foo", "dc1")
	servicesList, _ := a.CatalogServiceEndpoints("foo", "bar")

	assert.Equal(t, 2, len(servicesList))
	assert.Equal(t, true, containsService("foo", servicesList))
	assert.Equal(t, true, containsService("bar", servicesList))
}

func containsService(serviceName string, serviceList [][]*api.CatalogService) bool {
	for _, services := range serviceList {
		for _, service := range services {
			if service.ServiceName == serviceName {
				return true
			}
		}
	}
	return false
}
