package agent

import (
	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/hashicorp/consul/api"
)

type agent struct {
	host       string
	token      string
	datacenter string
}

//ConsulAgent describes consul agent behaviour
type ConsulAgent interface {
	Locality() *cp.Locality
	CatalogServiceEndpoints(serviceName string) ([]*api.CatalogService, error)
	WatchParams() map[string]string
}

//Catalog configures the consul agent catalog service client with the settings from current Agent.
func (a agent) catalog() (*api.Catalog, error) {
	config := api.DefaultConfig()
	config.Address = a.host
	config.Token = a.token
	config.Datacenter = a.datacenter
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return client.Catalog(), nil
}

//Locality translates Agent info to envoy control plane locality
func (a agent) Locality() *cp.Locality {
	return &cp.Locality{
		Region: a.datacenter,
	}
}

//CatalogServiceEndpoints makes an api call to consul agent host and gets list of catalog services
func (a agent) CatalogServiceEndpoints(serviceName string) ([]*api.CatalogService, error) {
	catalog, err := a.catalog()
	if err != nil {
		return nil, err
	}
	services, _, err := catalog.Service(serviceName, "", nil)
	return services, err
}

func (a agent) WatchParams() map[string]string {
	return map[string]string{"datacenter": a.datacenter, "token": a.token}
}

//NewAgent creates a new instance of a ConsulAgent
func NewAgent(host, token, datacenter string) ConsulAgent {
	return agent{host: host, token: token, datacenter: datacenter}
}
