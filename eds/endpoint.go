package eds

import (
	"time"

	"github.com/gojektech/consul-envoy-xds/agent"
	"github.com/gojektech/consul-envoy-xds/pubsub"

	"log"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	cpcore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	eds "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/watch"
)

//Endpoint is an agent catalog service
type Endpoint interface {
	Clusters() []*cp.Cluster
	Routes() []*cp.RouteConfiguration
	CLA() []*cp.ClusterLoadAssignment
	WatchPlan(publish func(*pubsub.Event)) (*watch.Plan, error)
}

type service struct {
	services map[string]Service
	agent    agent.ConsulAgent
}

func (s *service) serviceNames() []string {
	names := []string{}
	for k, _ := range s.services {
		names = append(names, k)
	}
	return names
}

func (s *service) getLbEndpoints(services []*api.CatalogService) []eds.LbEndpoint {
	var hosts []eds.LbEndpoint
	for _, s := range services {
		hosts = append(hosts, NewServiceHost(s).LbEndpoint())
	}
	return hosts
}

func (s *service) CLA() []*cp.ClusterLoadAssignment {
	serviceList, _ := s.agent.CatalogServiceEndpoints(s.serviceNames()...)
	log.Printf("discovered services from consul catalog for EDS: %v", serviceList)

	var cLAs []*cp.ClusterLoadAssignment
	for _, services := range serviceList {
		if len(services) > 0 {
			cLAs = append(cLAs, &cp.ClusterLoadAssignment{
				ClusterName: services[0].ServiceName,
				Policy: &cp.ClusterLoadAssignment_Policy{
					DropOverload: 0.0,
				},
				Endpoints: []eds.LocalityLbEndpoints{{
					Locality:    s.agent.Locality(),
					LbEndpoints: s.getLbEndpoints(services),
				}},
			})
		}
	}
	return cLAs
}

func (s *service) Clusters() []*cp.Cluster {
	serviceList, _ := s.agent.CatalogServiceEndpoints(s.serviceNames()...)
	log.Printf("discovered services from consul catalog for CDS: %v", serviceList)

	var clusters []*cp.Cluster
	for _, services := range serviceList {
		if len(services) > 0 {
			clusters = append(clusters, &cp.Cluster{
				Name:              services[0].ServiceName,
				Type:              cp.Cluster_EDS,
				ConnectTimeout:    1 * time.Second,
				ProtocolSelection: cp.Cluster_USE_DOWNSTREAM_PROTOCOL,
				EdsClusterConfig: &cp.Cluster_EdsClusterConfig{
					EdsConfig: &cpcore.ConfigSource{
						ConfigSourceSpecifier: &cpcore.ConfigSource_Ads{
							Ads: &cpcore.AggregatedConfigSource{},
						},
					},
				},
			})
		}
	}
	return clusters
}

func (s *service) Routes() []*cp.RouteConfiguration {
	serviceList, _ := s.agent.CatalogServiceEndpoints(s.serviceNames()...)
	log.Printf("discovered services from consul catalog for RDS: %v", serviceList)

	var routes []route.Route
	for _, services := range serviceList {
		if len(services) > 0 {
			name := services[0].ServiceName
			routes = append(routes, getRoutes(name, s.services[name].Whitelist)...)
		}
	}
	routeConfig := &cp.RouteConfiguration{
		Name: "local_route",
		VirtualHosts: []route.VirtualHost{{
			Name:    "local_service",
			Domains: []string{"*"},
			Routes:  routes,
		}},
	}
	return []*cp.RouteConfiguration{routeConfig}
}

func (s *service) WatchPlan(publish func(*pubsub.Event)) (*watch.Plan, error) {
	plan, err := watch.Parse(map[string]interface{}{
		"type":       "services",
		"datacenter": s.agent.WatchParams()["datacenter"],
		"token":      s.agent.WatchParams()["token"],
	})

	if err != nil {
		return nil, err
	}
	plan.Handler = func(idx uint64, data interface{}) {
		log.Println("consul watch triggerred")
		publish(&pubsub.Event{CLA: s.CLA(), Clusters: s.Clusters(), Routes: s.Routes()})
	}
	return plan, nil
}

func getRoutes(cluster string, pathPrefixes []string) []route.Route {
	var routes []route.Route
	for _, pathPrefix := range pathPrefixes {
		routes = append(routes, route.Route{
			Match: route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{
					Prefix: pathPrefix,
				},
			},
			Action: &route.Route_Route{
				Route: &route.RouteAction{
					ClusterSpecifier: &route.RouteAction_Cluster{
						Cluster: cluster,
					},
				},
			},
		})
	}
	return routes
}

//NewEndpoint creates an ServiceEndpoint representation
func NewEndpoint(services []Service, a agent.ConsulAgent) Endpoint {
	svcMap := map[string]Service{}
	for _, s := range services {
		svcMap[s.Name] = s
	}
	return &service{
		services: svcMap,
		agent:    a,
	}
}
