package eds

import (
	"fmt"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/cluster"
	"github.com/gojek/consul-envoy-xds/utils"
	"time"

	"github.com/gojek/consul-envoy-xds/agent"
	"github.com/gojek/consul-envoy-xds/pubsub"

	"log"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	cpcore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	eds "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	"github.com/gojek/consul-envoy-xds/config"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/watch"
	"strings"
)

//Endpoint is an agent catalog service
type Endpoint interface {
	Clusters() []*cp.Cluster
	Routes() []*cp.RouteConfiguration
	CLA() []*cp.ClusterLoadAssignment
	WatchPlan(publish func(*pubsub.Event)) (*watch.Plan, error)
}

const (
	regexPathIdentifier = "%regex:"
)

type service struct {
	services            map[string]Service
	agent               agent.ConsulAgent
	httpRateLimitConfig *config.HTTPHeaderRateLimitConfig
	enableHealthCheck   bool
}

var _ Endpoint = &service{}

func (s *service) serviceNames() []string {
	var names []string
	for k := range s.services {
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
	serviceList, _ := s.getCatalogServiceEndpoints(s.enableHealthCheck)
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
	serviceList, _ := s.getCatalogServiceEndpoints(s.enableHealthCheck)
	log.Printf("discovered services from consul catalog for CDS: %v", serviceList)

	var clusters []*cp.Cluster
	for _, services := range serviceList {
		if len(services) > 0 {
			serviceName := services[0].ServiceName
			svc := s.services[serviceName]

			clusters = append(clusters, &cp.Cluster{
				Name:              serviceName,
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
				CircuitBreakers: &cluster.CircuitBreakers{
					Thresholds: []*cluster.CircuitBreakers_Thresholds{{
						cpcore.RoutingPriority_DEFAULT,
						utils.Uint32Value(svc.CircuirBreakerConfig.MaxConnections),
						utils.Uint32Value(svc.CircuirBreakerConfig.MaxPendingRequests),
						utils.Uint32Value(svc.CircuirBreakerConfig.MaxRequests),
						utils.Uint32Value(svc.CircuirBreakerConfig.MaxRetries),
					}},
				},
			})
		}
	}
	return clusters
}

func (s *service) Routes() []*cp.RouteConfiguration {
	var routes []route.Route
	for _, serviceName := range s.serviceNames() {
		routes = append(routes, getRoutes(serviceName, s.services[serviceName].Whitelist)...)
	}
	routeConfig := &cp.RouteConfiguration{
		Name: "local_route",
		VirtualHosts: []route.VirtualHost{{
			Name:       "local_service",
			Domains:    []string{"*"},
			Routes:     routes,
			RateLimits: getRateLimits(s.httpRateLimitConfig),
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
		log.Println(fmt.Sprintf("consul watch triggerred: %v", data))
		publish(&pubsub.Event{CLA: s.CLA(), Clusters: s.Clusters(), Routes: s.Routes()})
	}
	return plan, nil
}

func (s *service) getCatalogServiceEndpoints(enableHealthCheck bool) ([][]*api.CatalogService, error) {
	if enableHealthCheck {
		healthCheckCatalogSvc, err := s.agent.HealthCheckCatalogServiceEndpoints(s.serviceNames()...)
		return healthCheckCatalogSvc, err
	}
	catalogSvc, err := s.agent.CatalogServiceEndpoints(s.serviceNames()...)
	return catalogSvc, err
}

func getRoutes(cluster string, whitelistPaths []string) []route.Route {
	var routes []route.Route
	for _, whitelistPath := range whitelistPaths {
		routes = append(routes, route.Route{
			Match: getPathSpecifier(whitelistPath),
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

func getPathSpecifier(whitelistPath string) route.RouteMatch {
	if strings.HasPrefix(whitelistPath, regexPathIdentifier) {
		regexPath := strings.Replace(whitelistPath, regexPathIdentifier, "", -1)
		return route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Regex{
				Regex: regexPath,
			},
		}
	}
	return route.RouteMatch{
		PathSpecifier: &route.RouteMatch_Prefix{
			Prefix: whitelistPath,
		}}
}

func getRateLimits(httpRateLimitConfig *config.HTTPHeaderRateLimitConfig) []*route.RateLimit {
	var rateLimits []*route.RateLimit

	if httpRateLimitConfig.IsEnabled {
		rateLimits = append(rateLimits, getHTTPHeaderRateLimit(httpRateLimitConfig))
	}

	return rateLimits
}

func getHTTPHeaderRateLimit(httpRateLimitConfig *config.HTTPHeaderRateLimitConfig) *route.RateLimit {
	actionSpecifier := route.RateLimit_Action_RequestHeaders_{
		RequestHeaders: &route.RateLimit_Action_RequestHeaders{
			HeaderName:    httpRateLimitConfig.HeaderName,
			DescriptorKey: httpRateLimitConfig.DescriptorKey,
		},
	}

	routeAction := route.RateLimit_Action{
		ActionSpecifier: &actionSpecifier,
	}

	return &route.RateLimit{
		Actions: []*route.RateLimit_Action{&routeAction},
	}
}

//NewEndpoint creates an ServiceEndpoint representation
func NewEndpoint(services []Service, a agent.ConsulAgent, httpRateLimitConfig *config.HTTPHeaderRateLimitConfig, enableHealthCheck bool) Endpoint {
	svcMap := map[string]Service{}
	for _, s := range services {
		svcMap[s.Name] = s
	}
	return &service{
		services:            svcMap,
		agent:               a,
		httpRateLimitConfig: httpRateLimitConfig,
		enableHealthCheck:   enableHealthCheck,
	}
}
