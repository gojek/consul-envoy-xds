package eds

import (
	"github.com/gojektech/consul-envoy-xds/agent"
	"github.com/gojektech/consul-envoy-xds/pubsub"

	cp "github.com/envoyproxy/go-control-plane/api"
	"github.com/hashicorp/consul/watch"
)

//Endpoint is an agent catalog service
type Endpoint interface {
	CLA() *cp.ClusterLoadAssignment
	WatchPlan(publish func(*pubsub.Event)) (*watch.Plan, error)
}

type service struct {
	name  string
	agent agent.ConsulAgent
}

func (s *service) getLbEndpoints() []*cp.LbEndpoint {
	hosts := make([]*cp.LbEndpoint, 0)
	services, _ := s.agent.CatalogServiceEndpoints(s.name)
	for _, s := range services {
		hosts = append(hosts, NewServiceHost(s).LbEndpoint())
	}
	return hosts
}

func (s *service) getLocalityEndpoints() []*cp.LocalityLbEndpoints {
	return []*cp.LocalityLbEndpoints{{Locality: s.agent.Locality(), LbEndpoints: s.getLbEndpoints()}}
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
		publish(&pubsub.Event{s.CLA(), &cp.Cluster{}})
	}
	return plan, nil
}

//NewEndpoint creates an ServiceEndpoint representation
func NewEndpoint(name string, a agent.ConsulAgent) Endpoint {
	return &service{name: name, agent: a}
}
