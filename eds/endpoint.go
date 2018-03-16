package eds

import (
	"fmt"
	"time"

	"github.com/gojektech/consul-envoy-xds/agent"
	"github.com/gojektech/consul-envoy-xds/pubsub"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	cpcore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	eds "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	"github.com/hashicorp/consul/watch"
)

//Endpoint is an agent catalog service
type Endpoint interface {
	Clusters() []*cp.Cluster
	Routes() []*cp.RouteConfiguration
	CLA() *cp.ClusterLoadAssignment
	WatchPlan(publish func(*pubsub.Event)) (*watch.Plan, error)
}

type service struct {
	name  string
	agent agent.ConsulAgent
}

func (s *service) getLbEndpoints() []eds.LbEndpoint {
	hosts := make([]eds.LbEndpoint, 0)
	services, _ := s.agent.CatalogServiceEndpoints(s.name)
	for _, s := range services {
		hosts = append(hosts, NewServiceHost(s).LbEndpoint())
	}
	return hosts
}

func (s *service) getLocalityEndpoints() []eds.LocalityLbEndpoints {
	return []eds.LocalityLbEndpoints{{Locality: s.agent.Locality(), LbEndpoints: s.getLbEndpoints()}}
}

func (s *service) claPolicy() *cp.ClusterLoadAssignment_Policy {
	return &cp.ClusterLoadAssignment_Policy{DropOverload: 0.0}
}

func (s *service) clusterName() string {
	return s.name
}

func (s *service) CLA() *cp.ClusterLoadAssignment {
	return &cp.ClusterLoadAssignment{Endpoints: s.getLocalityEndpoints(), ClusterName: s.clusterName(), Policy: s.claPolicy()}
}

func (s *service) Clusters() []*cp.Cluster {
	services, _ := s.agent.CatalogServiceEndpoints(s.name)
	if len(services) > 0 {
		return []*cp.Cluster{&cp.Cluster{
			Name:              fmt.Sprintf("%s", services[0].ServiceName),
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
		}}
	}
	return []*cp.Cluster{}
}

func (s *service) Routes() []*cp.RouteConfiguration {
	services, _ := s.agent.CatalogServiceEndpoints(s.name)
	if len(services) > 0 {
		return []*cp.RouteConfiguration{&cp.RouteConfiguration{
			Name: "local_route",
			VirtualHosts: []route.VirtualHost{route.VirtualHost{
				Name:    "local_service",
				Domains: []string{"*"},
				Routes: []route.Route{route.Route{
					Match: route.RouteMatch{
						PathSpecifier: &route.RouteMatch_Prefix{
							Prefix: "/",
						},
					},
					Action: &route.Route_Route{
						Route: &route.RouteAction{
							ClusterSpecifier: &route.RouteAction_Cluster{
								Cluster: fmt.Sprintf("%s", services[0].ServiceName),
							},
						},
					},
				}},
			}},
		}}
	}
	return []*cp.RouteConfiguration{}
}

func (s *service) WatchPlan(publish func(*pubsub.Event)) (*watch.Plan, error) {
	plan, err := watch.Parse(map[string]interface{}{
		"type":       "service",
		"service":    s.clusterName(),
		"datacenter": s.agent.WatchParams()["datacenter"],
		"token":      s.agent.WatchParams()["token"],
	})

	if err != nil {
		return nil, err
	}
	plan.Handler = func(idx uint64, data interface{}) {
		println("consul watch triggerred")
		publish(&pubsub.Event{s.CLA(), s.Clusters(), s.Routes()})
	}
	return plan, nil
}

//NewEndpoint creates an ServiceEndpoint representation
func NewEndpoint(name string, a agent.ConsulAgent) Endpoint {
	return &service{name: name, agent: a}
}
