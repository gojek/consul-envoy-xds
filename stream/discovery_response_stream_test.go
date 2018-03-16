package stream_test

import (
	"testing"

	"github.com/gojektech/consul-envoy-xds/stream"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShouldShouldStreamEndpointDiscoveryResponse(t *testing.T) {
	mockStream := &stream.MockXDSStream{}
	mockStream.On("Send", mock.AnythingOfType("*v2.DiscoveryResponse")).Return(nil)
	drs := stream.NewDiscoveryResponseStream(mockStream)
	cla := &cp.ClusterLoadAssignment{}
	drs.SendEDS(cla)
	assert.Equal(t, "type.googleapis.com/envoy.api.v2.ClusterLoadAssignment", mockStream.Capture().TypeUrl)
	capturedResponse := mockStream.Capture()
	assert.Equal(t, 1, len(capturedResponse.Resources))
	assert.Equal(t, "type.googleapis.com/envoy.api.v2.ClusterLoadAssignment", capturedResponse.Resources[0].TypeUrl)
	claBytes, _ := proto.Marshal(cla)
	assert.Equal(t, claBytes, capturedResponse.Resources[0].Value)

	mockStream.AssertExpectations(t)
}

func TestShouldShouldStreamEndpointDiscoveryResponseWithIncrementingNonceAndVersion(t *testing.T) {
	mockStream := &stream.MockXDSStream{}
	mockStream.On("Send", mock.AnythingOfType("*v2.DiscoveryResponse")).Times(3).Return(nil)
	drs := stream.NewDiscoveryResponseStream(mockStream)
	drs.SendEDS(&cp.ClusterLoadAssignment{})
	assert.Equal(t, "0", mockStream.Capture().Nonce)
	assert.Equal(t, "0", mockStream.Capture().VersionInfo)
	drs.SendEDS(&cp.ClusterLoadAssignment{})
	assert.Equal(t, "1", mockStream.Capture().Nonce)
	assert.Equal(t, "1", mockStream.Capture().VersionInfo)
	drs.SendEDS(&cp.ClusterLoadAssignment{})
	assert.Equal(t, "2", mockStream.Capture().Nonce)
	assert.Equal(t, "2", mockStream.Capture().VersionInfo)
	mockStream.AssertExpectations(t)
}

func TestShouldShouldStreamClusterDiscoveryResponse(t *testing.T) {
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

func TestShouldShouldStreamClusterDiscoveryResponseWithIncrementingNonceAndVersion(t *testing.T) {
	mockStream := &stream.MockXDSStream{}
	mockStream.On("Send", mock.AnythingOfType("*v2.DiscoveryResponse")).Times(3).Return(nil)
	drs := stream.NewDiscoveryResponseStream(mockStream)
	drs.SendCDS([]*cp.Cluster{&cp.Cluster{}})
	assert.Equal(t, "0", mockStream.Capture().Nonce)
	assert.Equal(t, "0", mockStream.Capture().VersionInfo)
	drs.SendCDS([]*cp.Cluster{&cp.Cluster{}})
	assert.Equal(t, "1", mockStream.Capture().Nonce)
	assert.Equal(t, "1", mockStream.Capture().VersionInfo)
	drs.SendCDS([]*cp.Cluster{&cp.Cluster{}})
	assert.Equal(t, "2", mockStream.Capture().Nonce)
	assert.Equal(t, "2", mockStream.Capture().VersionInfo)
	mockStream.AssertExpectations(t)
}

func TestShouldShouldStreamRouteConfigurationResponse(t *testing.T) {
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

func TestShouldShouldStreamRouteConfigurationResponseWithIncrementingNonceAndVersion(t *testing.T) {
	mockStream := &stream.MockXDSStream{}
	mockStream.On("Send", mock.AnythingOfType("*v2.DiscoveryResponse")).Times(3).Return(nil)
	drs := stream.NewDiscoveryResponseStream(mockStream)
	drs.SendRDS([]*cp.RouteConfiguration{&cp.RouteConfiguration{}})
	assert.Equal(t, "0", mockStream.Capture().Nonce)
	assert.Equal(t, "0", mockStream.Capture().VersionInfo)
	drs.SendRDS([]*cp.RouteConfiguration{&cp.RouteConfiguration{}})
	assert.Equal(t, "1", mockStream.Capture().Nonce)
	assert.Equal(t, "1", mockStream.Capture().VersionInfo)
	drs.SendRDS([]*cp.RouteConfiguration{&cp.RouteConfiguration{}})
	assert.Equal(t, "2", mockStream.Capture().Nonce)
	assert.Equal(t, "2", mockStream.Capture().VersionInfo)
	mockStream.AssertExpectations(t)
}
