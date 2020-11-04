package app_test

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	cpcore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	eds "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	dis "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	google_protobuf5 "github.com/gogo/protobuf/types"
	"github.com/gojek/consul-envoy-xds/agent"
	"github.com/gojek/consul-envoy-xds/app"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/lib/freeport"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestPushToEnvoyWhenConsulWatchTriggers(t *testing.T) {
	ports, _ := freeport.Free(1)
	xdsPort := ports[0]

	consulSvr, consulClient := agent.StartConsulTestServer()
	defer consulSvr.Stop()

	os.Setenv("PORT", strconv.Itoa(xdsPort))
	defer os.Unsetenv("PORT")
	_, port, _ := net.SplitHostPort(consulSvr.HTTPAddr)
	os.Setenv("CONSUL_CLIENT_PORT", port)
	defer os.Unsetenv("CONSUL_CLIENT_PORT")
	os.Setenv("WATCHED_SERVICE", "testSvc1,testSvc2")
	defer os.Unsetenv("WATCHED_SERVICE")
	os.Setenv("TESTSVC1_WHITELISTED_ROUTES", "/foo,%regex:/bar")
	defer os.Unsetenv("TESTSVC1_WHITELISTED_ROUTES")
	os.Setenv("TESTSVC2_WHITELISTED_ROUTES", "%regex:/hoo,/car")
	defer os.Unsetenv("TESTSVC2_WHITELISTED_ROUTES")

	go app.Start()
	defer app.Stop()
	waitForAppToStart(xdsPort)

	conn, _ := grpc.Dial("localhost:"+strconv.Itoa(xdsPort), grpc.WithInsecure())
	defer conn.Close()
	xDSClient := dis.NewAggregatedDiscoveryServiceClient(conn)
	stream, err := xDSClient.StreamAggregatedResources(context.Background())
	assert.NoError(t, err)

	var messages []google_protobuf5.Any
	go func() {
		for {
			res, err := stream.Recv()
			if err != nil {
				return
			}
			messages = append(messages, res.GetResources()...)
		}
	}()

	testSvc1 := startTestSvc(consulClient, "testSvc1")
	defer testSvc1.Close()

	testSvc2 := startTestSvc(consulClient, "testSvc2")
	defer testSvc2.Close()

	assertCluster(t, &messages, "testSvc1")
	assertCLA(t, &messages, "testSvc1", testSvc1.Listener.Addr().(*net.TCPAddr).Port)
	assertCluster(t, &messages, "testSvc2")
	assertCLA(t, &messages, "testSvc2", testSvc2.Listener.Addr().(*net.TCPAddr).Port)
	expectedRoutes := []route.Route{
		routeWithPathPrefix("testSvc1", "/foo"),
		routeWithRegexPath("testSvc1", "/bar"),
		routeWithRegexPath("testSvc2", "/hoo"),
		routeWithPathPrefix("testSvc2", "/car"),
	}
	assertRouteConfig(t, messages, expectedRoutes)
}

func TestRespondToEnvoyOnRequest(t *testing.T) {
	ports, _ := freeport.Free(1)
	xdsPort := ports[0]

	consulSvr, consulClient := agent.StartConsulTestServer()
	defer consulSvr.Stop()

	os.Setenv("PORT", strconv.Itoa(xdsPort))
	defer os.Unsetenv("PORT")
	_, port, _ := net.SplitHostPort(consulSvr.HTTPAddr)
	os.Setenv("CONSUL_CLIENT_PORT", port)
	defer os.Unsetenv("CONSUL_CLIENT_PORT")
	os.Setenv("WATCHED_SERVICE", "testSvc1")
	defer os.Unsetenv("WATCHED_SERVICE")

	go app.Start()
	defer app.Stop()
	waitForAppToStart(xdsPort)

	conn, _ := grpc.Dial("localhost:"+strconv.Itoa(xdsPort), grpc.WithInsecure())
	defer conn.Close()
	xDSClient := dis.NewAggregatedDiscoveryServiceClient(conn)
	stream, err := xDSClient.StreamAggregatedResources(context.Background())
	assert.NoError(t, err)

	var messages []google_protobuf5.Any
	go func() {
		for {
			res, err := stream.Recv()
			if err != nil {
				return
			}
			messages = append(messages, res.GetResources()...)
		}
	}()

	testSvc1 := startTestSvc(consulClient, "testSvc1")
	defer testSvc1.Close()

	stream.Send(&cp.DiscoveryRequest{})
	stream.CloseSend()

	assertCluster(t, &messages, "testSvc1")
	assertCLA(t, &messages, "testSvc1", testSvc1.Listener.Addr().(*net.TCPAddr).Port)
	assertRouteConfig(t, messages, []route.Route{routeWithPathPrefix("testSvc1", "/")})
}

func assertRouteConfig(t *testing.T, messages []google_protobuf5.Any, routes []route.Route) {
	routeConfig, _ := google_protobuf5.MarshalAny(&cp.RouteConfiguration{
		Name: "local_route",
		VirtualHosts: []route.VirtualHost{{
			Name:    "local_service",
			Domains: []string{"*"},
			Routes:  routes,
		}},
	})
	var publishedMessage google_protobuf5.Any
	for _, message := range messages {
		if message.TypeUrl == "type.googleapis.com/envoy.api.v2.RouteConfiguration" {
			publishedMessage = message
		}
	}
	assert.Equal(t, publishedMessage.Compare(routeConfig), 0)
	assert.Contains(t, messages, *routeConfig)
}

func routeWithPathPrefix(name string, pathPrefix string) route.Route {
	return route.Route{
		Match: route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: pathPrefix,
			},
		},
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: name,
				},
			},
		},
	}
}

func routeWithRegexPath(name string, pathPrefix string) route.Route {
	return route.Route{
		Match: route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Regex{
				Regex: pathPrefix,
			},
		},
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: name,
				},
			},
		},
	}
}

func assertCLA(t *testing.T, messages *[]google_protobuf5.Any, cluster string, port int) bool {
	cla, _ := google_protobuf5.MarshalAny(&cp.ClusterLoadAssignment{Endpoints: []eds.LocalityLbEndpoints{{
		Locality: &cpcore.Locality{
			Region: "dc1",
		},
		LbEndpoints: []eds.LbEndpoint{{
			HealthStatus: cpcore.HealthStatus_HEALTHY,
			Endpoint: &eds.Endpoint{
				Address: &cpcore.Address{
					Address: &cpcore.Address_SocketAddress{
						SocketAddress: &cpcore.SocketAddress{
							Protocol: cpcore.TCP,
							Address:  "localhost",
							PortSpecifier: &cpcore.SocketAddress_PortValue{
								PortValue: uint32(port),
							},
						},
					},
				}}}},
	}}, ClusterName: cluster, Policy: &cp.ClusterLoadAssignment_Policy{DropOverload: 0.0}})
	await(func() bool {
		return IncludeElement(*messages, *cla)
	})
	return assert.Contains(t, *messages, *cla)
}

func assertCluster(t *testing.T, messages *[]google_protobuf5.Any, name string) bool {
	cluster, _ := google_protobuf5.MarshalAny(&cp.Cluster{
		Name:              name,
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
	await(func() bool {
		return IncludeElement(*messages, *cluster)
	})
	return assert.Contains(t, *messages, *cluster)
}

func startTestSvc(consulClient *consulapi.Client, name string) *httptest.Server {
	testSvc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	consulClient.Agent().ServiceRegister(&consulapi.AgentServiceRegistration{
		Name:    name,
		ID:      name,
		Address: "localhost",
		Port:    testSvc.Listener.Addr().(*net.TCPAddr).Port,
	})
	return testSvc
}

func waitForAppToStart(port int) {
	retry(100, 10*time.Millisecond, func() error {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(port), 1*time.Millisecond)
		if err == nil {
			defer conn.Close()
		}
		return err
	})
}

func retry(attempts int, sleep time.Duration, fn func() error) error {
	if err := fn(); err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			return retry(attempts, sleep, fn)
		}
		return err
	}
	return nil
}

func IncludeElement(list interface{}, element interface{}) (found bool) {

	listValue := reflect.ValueOf(list)
	elementValue := reflect.ValueOf(element)

	if reflect.TypeOf(list).Kind() == reflect.String {
		return strings.Contains(listValue.String(), elementValue.String())
	}

	if reflect.TypeOf(list).Kind() == reflect.Map {
		mapKeys := listValue.MapKeys()
		for i := 0; i < len(mapKeys); i++ {
			if assert.ObjectsAreEqual(mapKeys[i].Interface(), element) {
				return true
			}
		}
		return false
	}

	for i := 0; i < listValue.Len(); i++ {
		if assert.ObjectsAreEqual(listValue.Index(i).Interface(), element) {
			return true
		}
	}
	return false
}

func await(f func() bool) bool {
	for i := 0; i < 1000; i++ {
		if f() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	log.Println("timeout waiting for condition")
	return false
}
