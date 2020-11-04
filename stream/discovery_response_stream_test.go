package stream_test

import (
	"testing"

	"github.com/gojek/consul-envoy-xds/stream"

	"strconv"
	"time"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShouldStreamEndpointDiscoveryResponse(t *testing.T) {
	mockStream := &stream.MockXDSStream{}
	mockStream.On("Send", mock.AnythingOfType("*v2.DiscoveryResponse")).Return(nil)
	drs := stream.NewDiscoveryResponseStream(mockStream)
	cla := []*cp.ClusterLoadAssignment{{}}
	drs.SendEDS(cla)
	assert.Equal(t, "type.googleapis.com/envoy.api.v2.ClusterLoadAssignment", mockStream.Capture().TypeUrl)
	capturedResponse := mockStream.Capture()
	assert.Equal(t, 1, len(capturedResponse.Resources))
	assert.Equal(t, "type.googleapis.com/envoy.api.v2.ClusterLoadAssignment", capturedResponse.Resources[0].TypeUrl)
	claBytes, _ := proto.Marshal(cla[0])
	assert.Equal(t, claBytes, capturedResponse.Resources[0].Value)

	mockStream.AssertExpectations(t)
}

func TestShouldStreamEndpointDiscoveryResponseWithMultipleClusterResources(t *testing.T) {
	mockStream := &stream.MockXDSStream{}
	mockStream.On("Send", mock.AnythingOfType("*v2.DiscoveryResponse")).Return(nil)
	drs := stream.NewDiscoveryResponseStream(mockStream)
	clusters := []*cp.Cluster{
		&cp.Cluster{Name: "foo"},
		&cp.Cluster{Name: "bar"},
	}
	drs.SendCDS(clusters)
	assert.Equal(t, "type.googleapis.com/envoy.api.v2.Cluster", mockStream.Capture().TypeUrl)
	capturedResponse := mockStream.Capture()
	assert.Equal(t, 2, len(capturedResponse.Resources))
	assert.Equal(t, "type.googleapis.com/envoy.api.v2.Cluster", capturedResponse.Resources[0].TypeUrl)
	assert.Equal(t, "type.googleapis.com/envoy.api.v2.Cluster", capturedResponse.Resources[1].TypeUrl)
	cluster1Bytes, _ := proto.Marshal(clusters[0])
	cluster2Bytes, _ := proto.Marshal(clusters[1])
	assert.Equal(t, cluster1Bytes, capturedResponse.Resources[0].Value)
	assert.Equal(t, cluster2Bytes, capturedResponse.Resources[1].Value)

	mockStream.AssertExpectations(t)
}

func TestShouldStreamEndpointDiscoveryResponseWithMultipleRouteResources(t *testing.T) {
	mockStream := &stream.MockXDSStream{}
	mockStream.On("Send", mock.AnythingOfType("*v2.DiscoveryResponse")).Return(nil)
	drs := stream.NewDiscoveryResponseStream(mockStream)
	routeConfig := []*cp.RouteConfiguration{
		&cp.RouteConfiguration{Name: "foo"},
		&cp.RouteConfiguration{Name: "bar"},
	}
	drs.SendRDS(routeConfig)
	assert.Equal(t, "type.googleapis.com/envoy.api.v2.RouteConfiguration", mockStream.Capture().TypeUrl)
	capturedResponse := mockStream.Capture()
	assert.Equal(t, 2, len(capturedResponse.Resources))
	assert.Equal(t, "type.googleapis.com/envoy.api.v2.RouteConfiguration", capturedResponse.Resources[0].TypeUrl)
	assert.Equal(t, "type.googleapis.com/envoy.api.v2.RouteConfiguration", capturedResponse.Resources[1].TypeUrl)
	route1Bytes, _ := proto.Marshal(routeConfig[0])
	route2Bytes, _ := proto.Marshal(routeConfig[1])
	assert.Equal(t, route1Bytes, capturedResponse.Resources[0].Value)
	assert.Equal(t, route2Bytes, capturedResponse.Resources[1].Value)

	mockStream.AssertExpectations(t)
}

func TestShouldStreamEndpointDiscoveryResponseWithIncrementingNonceAndVersion(t *testing.T) {
	mockStream := &stream.MockXDSStream{}
	now := time.Now().UnixNano()
	mockStream.On("Send", mock.AnythingOfType("*v2.DiscoveryResponse")).Times(3).Return(nil)

	drs := stream.NewDiscoveryResponseStream(mockStream)
	drs.SendEDS([]*cp.ClusterLoadAssignment{})
	nonce1, _ := strconv.Atoi(mockStream.Capture().Nonce)
	assert.True(t, int64(nonce1) > now)
	version1, _ := strconv.Atoi(mockStream.Capture().VersionInfo)
	assert.True(t, int64(version1) > now)

	drs.SendEDS([]*cp.ClusterLoadAssignment{})
	nonce2, _ := strconv.Atoi(mockStream.Capture().Nonce)
	assert.True(t, int64(nonce2) > now)
	version2, _ := strconv.Atoi(mockStream.Capture().VersionInfo)
	assert.True(t, int64(version2) > now)

	drs.SendEDS([]*cp.ClusterLoadAssignment{})
	nonce3, _ := strconv.Atoi(mockStream.Capture().Nonce)
	assert.True(t, int64(nonce3) > now)
	version3, _ := strconv.Atoi(mockStream.Capture().VersionInfo)
	assert.True(t, int64(version3) > now)
	assert.True(t, (version1 < version2) && (version2 < version3))
	assert.True(t, (nonce1 < nonce2) && (nonce2 < nonce3))
	mockStream.AssertExpectations(t)
}

func TestShouldStreamClusterDiscoveryResponse(t *testing.T) {
	mockStream := &stream.MockXDSStream{}
	mockStream.On("Send", mock.AnythingOfType("*v2.DiscoveryResponse")).Return(nil)
	drs := stream.NewDiscoveryResponseStream(mockStream)
	cluster := &cp.Cluster{}
	drs.SendCDS([]*cp.Cluster{cluster})
	assert.Equal(t, "type.googleapis.com/envoy.api.v2.Cluster", mockStream.Capture().TypeUrl)
	capturedResponse := mockStream.Capture()
	assert.Equal(t, 1, len(capturedResponse.Resources))
	assert.Equal(t, "type.googleapis.com/envoy.api.v2.Cluster", capturedResponse.Resources[0].TypeUrl)
	clusterBytes, _ := proto.Marshal(cluster)
	assert.Equal(t, clusterBytes, capturedResponse.Resources[0].Value)

	mockStream.AssertExpectations(t)
}

func TestShouldStreamClusterDiscoveryResponseWithIncrementingNonceAndVersion(t *testing.T) {
	mockStream := &stream.MockXDSStream{}
	now := time.Now().UnixNano()

	mockStream.On("Send", mock.AnythingOfType("*v2.DiscoveryResponse")).Times(3).Return(nil)
	drs := stream.NewDiscoveryResponseStream(mockStream)
	drs.SendCDS([]*cp.Cluster{{}})
	nonce1, _ := strconv.Atoi(mockStream.Capture().Nonce)
	assert.True(t, int64(nonce1) > now)
	version1, _ := strconv.Atoi(mockStream.Capture().VersionInfo)
	assert.True(t, int64(version1) > now)

	drs.SendCDS([]*cp.Cluster{{}})
	nonce2, _ := strconv.Atoi(mockStream.Capture().Nonce)
	assert.True(t, int64(nonce2) > now)
	version2, _ := strconv.Atoi(mockStream.Capture().VersionInfo)
	assert.True(t, int64(version2) > now)

	drs.SendCDS([]*cp.Cluster{{}})
	nonce3, _ := strconv.Atoi(mockStream.Capture().Nonce)
	assert.True(t, int64(nonce3) > now)
	version3, _ := strconv.Atoi(mockStream.Capture().VersionInfo)
	assert.True(t, int64(version3) > now)
	assert.True(t, (version1 < version2) && (version2 < version3))
	assert.True(t, (nonce1 < nonce2) && (nonce2 < nonce3))
	mockStream.AssertExpectations(t)
}

func TestShouldStreamRouteConfigurationResponse(t *testing.T) {
	mockStream := &stream.MockXDSStream{}
	mockStream.On("Send", mock.AnythingOfType("*v2.DiscoveryResponse")).Return(nil)
	drs := stream.NewDiscoveryResponseStream(mockStream)
	route := &cp.RouteConfiguration{}
	drs.SendRDS([]*cp.RouteConfiguration{route})
	assert.Equal(t, "type.googleapis.com/envoy.api.v2.RouteConfiguration", mockStream.Capture().TypeUrl)
	capturedResponse := mockStream.Capture()
	assert.Equal(t, 1, len(capturedResponse.Resources))
	assert.Equal(t, "type.googleapis.com/envoy.api.v2.RouteConfiguration", capturedResponse.Resources[0].TypeUrl)
	routeBytes, _ := proto.Marshal(route)
	assert.Equal(t, routeBytes, capturedResponse.Resources[0].Value)

	mockStream.AssertExpectations(t)
}

func TestShouldStreamRouteConfigurationResponseWithIncrementingNonceAndVersion(t *testing.T) {
	mockStream := &stream.MockXDSStream{}
	now := time.Now().UnixNano()

	mockStream.On("Send", mock.AnythingOfType("*v2.DiscoveryResponse")).Times(3).Return(nil)
	drs := stream.NewDiscoveryResponseStream(mockStream)
	drs.SendRDS([]*cp.RouteConfiguration{{}})
	nonce1, _ := strconv.Atoi(mockStream.Capture().Nonce)
	assert.True(t, int64(nonce1) > now)
	version1, _ := strconv.Atoi(mockStream.Capture().VersionInfo)
	assert.True(t, int64(version1) > now)

	drs.SendRDS([]*cp.RouteConfiguration{{}})
	nonce2, _ := strconv.Atoi(mockStream.Capture().Nonce)
	assert.True(t, int64(nonce2) > now)
	version2, _ := strconv.Atoi(mockStream.Capture().VersionInfo)
	assert.True(t, int64(version2) > now)

	drs.SendRDS([]*cp.RouteConfiguration{{}})
	nonce3, _ := strconv.Atoi(mockStream.Capture().Nonce)
	assert.True(t, int64(nonce3) > now)
	version3, _ := strconv.Atoi(mockStream.Capture().VersionInfo)
	assert.True(t, int64(version3) > now)
	assert.True(t, (version1 < version2) && (version2 < version3))
	assert.True(t, (nonce1 < nonce2) && (nonce2 < nonce3))
	mockStream.AssertExpectations(t)
}
