package agent_test

import (
	"fmt"
	"github.com/gojektech/consul-envoy-xds/agent"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/testutil/retry"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
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

func TestShouldReturnHealthCheckPassedServicesFromCatalog(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(w,"helo")
	}))
	defer ts.Close()

	consulSvr, consulClient := agent.StartConsulTestServer()
	defer consulSvr.Stop()

	consulClient.Agent().ServiceRegister(&api.AgentServiceRegistration{Name: "foo", ID: "foo", Check: &api.AgentServiceCheck{
		HTTP:      "invalidhost:8080",
		Interval: "1ms",
	}})
	consulClient.Agent().ServiceRegister(&api.AgentServiceRegistration{Name: "bar", ID: "bar", Check: &api.AgentServiceCheck{
		HTTP:      ts.URL,
		Interval: "1ms",
	}})

	a := agent.NewAgent(consulSvr.HTTPAddr, "token", "dc1")

	retry.Run(t, func(r *retry.R) {
		servicesList, _ := a.HealthCheckCatalogServiceEndpoints("foo", "bar")
		if len(servicesList) == 0 {
			r.Fatalf("Bad: %v", servicesList)
		}
		assert.Equal(t, 1, len(servicesList))
		assert.Equal(t, false, containsService("foo", servicesList))
		assert.Equal(t, true, containsService("bar", servicesList))
	})
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
