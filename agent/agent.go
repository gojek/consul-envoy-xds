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
	CatalogServiceEndpoints(serviceName ...string) ([][]*api.CatalogService, error)
	HealthCheckCatalogServiceEndpoints(serviceNames ...string) ([][]*api.CatalogService, error)
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

//Health configures the consul agent health service client with the settings from current Agent.
func (a agent) health() (*api.Health, error) {
	config := api.DefaultConfig()
	config.Address = a.host
	config.Token = a.token
	config.Datacenter = a.datacenter
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return client.Health(), nil
}

//Locality translates Agent info to envoy control plane locality
func (a agent) Locality() *cp.Locality {
	return &cp.Locality{
		Region: a.datacenter,
	}
}

//CatalogServiceEndpoints makes an api call to consul agent host and gets list of catalog services
func (a agent) CatalogServiceEndpoints(serviceNames ...string) ([][]*api.CatalogService, error) {
	catalog, err := a.catalog()
	if err != nil {
		return nil, err
	}
	var services [][]*api.CatalogService
	for _, serviceName := range serviceNames {
		catalogSvc, _, err := catalog.Service(serviceName, "", nil)
		if err != nil {
			return services, err
		}
		services = append(services, catalogSvc)
	}
	return services, err
}

//HealthCheckCatalogServiceEndpoints makes an api call to consul agent host and gets list of health check passed catalog services only
func (a agent) HealthCheckCatalogServiceEndpoints(serviceNames ...string) ([][]*api.CatalogService, error) {
	health, err := a.health()
	if err != nil {
		return nil, err
	}
	var services [][]*api.CatalogService
	for _, serviceName := range serviceNames {
		healthSvc, _, err := health.Service(serviceName, "", true, nil)
		if err != nil {
			return services, err
		}
		if len(healthSvc) > 0 {
			catalogSvc := buildCatalogServicesFromServiceEntries(healthSvc)
			services = append(services, catalogSvc)
		}

	}
	return services, err
}

func (a agent) WatchParams() map[string]string {
	return map[string]string{"datacenter": a.datacenter, "token": a.token}
}

//NewAgent creates a new instance of a ConsulAgent
func NewAgent(host, token, datacenter string) ConsulAgent {
	return agent{host: host, token: token, datacenter: datacenter}
}

func buildCatalogServicesFromServiceEntries(ServiceEntries []*api.ServiceEntry) []*api.CatalogService {
	var catalogServices []*api.CatalogService
	for _, serviceEntry := range ServiceEntries {
		catalogSvc := &api.CatalogService{
			ID:                       serviceEntry.Node.ID,
			Node:                     serviceEntry.Node.Node,
			Address:                  serviceEntry.Node.Address,
			Datacenter:               serviceEntry.Node.Datacenter,
			TaggedAddresses:          serviceEntry.Node.TaggedAddresses,
			NodeMeta:                 serviceEntry.Node.Meta,
			ServiceID:                serviceEntry.Service.ID,
			ServiceName:              serviceEntry.Service.Service,
			ServiceAddress:           serviceEntry.Service.Address,
			ServiceTags:              serviceEntry.Service.Tags,
			ServiceMeta:              serviceEntry.Service.Meta,
			ServicePort:              serviceEntry.Service.Port,
			ServiceEnableTagOverride: serviceEntry.Service.EnableTagOverride,
			CreateIndex:              serviceEntry.Service.CreateIndex,
			ModifyIndex:              serviceEntry.Service.ModifyIndex,
		}
		catalogServices = append(catalogServices, catalogSvc)
	}
	return catalogServices
}
