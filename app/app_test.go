package app_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	cpcore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	eds "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	dis "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	google_protobuf5 "github.com/gogo/protobuf/types"
	"github.com/gojektech/consul-envoy-xds/agent"
	"github.com/gojektech/consul-envoy-xds/app"
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
	os.Setenv("TESTSVC1_WHITELISTED_ROUTES", "/foo,/bar")
	defer os.Unsetenv("TESTSVC1_WHITELISTED_ROUTES")
	os.Setenv("TESTSVC2_WHITELISTED_ROUTES", "/hoo,/car")
	defer os.Unsetenv("TESTSVC2_WHITELISTED_ROUTES")

	go app.Start()
	defer app.Stop()
	waitForAppToStart(xdsPort)

	testSvc1 := startTestSvc(consulClient, "testSvc1")
	defer testSvc1.Close()

	testSvc2 := startTestSvc(consulClient, "testSvc2")
	defer testSvc2.Close()

	conn, _ := grpc.Dial("localhost:"+strconv.Itoa(xdsPort), grpc.WithInsecure())
	defer conn.Close()
	xDSClient := dis.NewAggregatedDiscoveryServiceClient(conn)
	stream, err := xDSClient.StreamAggregatedResources(context.Background())
	assert.NoError(t, err)

	var messages []google_protobuf5.Any
	for i := 0; i < 3; i++ {
		res, err := stream.Recv()
		assert.NoError(t, err)
		messages = append(messages, res.GetResources()...)
	}

	assertCluster(t, messages, "testSvc1")
	assertCLA(t, messages, "testSvc1", testSvc1.Listener.Addr().(*net.TCPAddr).Port)

	assertCluster(t, messages, "testSvc2")
	assertCLA(t, messages, "testSvc2", testSvc2.Listener.Addr().(*net.TCPAddr).Port)

	assertRouteConfig(t, messages, []string{"testSvc1", "testSvc2"}, [][]string{[]string{"/foo", "/bar"}, []string{"/hoo", "/car"}})
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

	testSvc1 := startTestSvc(consulClient, "testSvc1")
	defer testSvc1.Close()

	conn, _ := grpc.Dial("localhost:"+strconv.Itoa(xdsPort), grpc.WithInsecure())
	defer conn.Close()
	xDSClient := dis.NewAggregatedDiscoveryServiceClient(conn)
	stream, err := xDSClient.StreamAggregatedResources(context.Background())
	assert.NoError(t, err)

	// flush incoming messages
	for i := 0; i < 3; i++ {
		stream.Recv()
	}

	stream.Send(&cp.DiscoveryRequest{})
	stream.CloseSend()

	var messages []google_protobuf5.Any
	for i := 0; i < 3; i++ {
		res, err := stream.Recv()
		assert.NoError(t, err)
		messages = append(messages, res.GetResources()...)
	}

	assertCluster(t, messages, "testSvc1")
	assertCLA(t, messages, "testSvc1", testSvc1.Listener.Addr().(*net.TCPAddr).Port)

	assertRouteConfig(t, messages, []string{"testSvc1"}, [][]string{[]string{"/"}})
}

func assertRouteConfig(t *testing.T, messages []google_protobuf5.Any, clusters []string, routeList [][]string) {
	var routes []route.Route
	for i, cluster := range clusters {
		for _, pathPrefix := range routeList[i] {
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
	}
	routeConfig, _ := google_protobuf5.MarshalAny(&cp.RouteConfiguration{
		Name: "local_route",
		VirtualHosts: []route.VirtualHost{{
			Name:    "local_service",
			Domains: []string{"*"},
			Routes:  routes,
		}},
	})
	assert.Contains(t, messages, *routeConfig)
}

func assertCLA(t *testing.T, messages []google_protobuf5.Any, cluster string, port int) {
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
	assert.Contains(t, messages, *cla)
}

func assertCluster(t *testing.T, messages []google_protobuf5.Any, name string) {
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
	assert.Contains(t, messages, *cluster)
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
