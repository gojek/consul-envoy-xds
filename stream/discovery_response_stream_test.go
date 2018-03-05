package stream_test

import (
	"testing"

	"github.com/gojektech/consul-envoy-xds/stream"

	cp "github.com/envoyproxy/go-control-plane/api"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShouldShouldStreamEndpointDiscoveryResponse(t *testing.T) {
	mockStream := &stream.MockXDSStream{}
	mockStream.On("Send", mock.AnythingOfType("*api.DiscoveryResponse")).Return(nil)
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
	mockStream.On("Send", mock.AnythingOfType("*api.DiscoveryResponse")).Times(3).Return(nil)
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
	mockStream.On("Send", mock.AnythingOfType("*api.DiscoveryResponse")).Return(nil)
	drs := stream.NewDiscoveryResponseStream(mockStream)
	cluster := &cp.Cluster{}
	drs.SendCDS(cluster)
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
	mockStream.On("Send", mock.AnythingOfType("*api.DiscoveryResponse")).Times(3).Return(nil)
	drs := stream.NewDiscoveryResponseStream(mockStream)
	drs.SendCDS(&cp.Cluster{})
	assert.Equal(t, "0", mockStream.Capture().Nonce)
	assert.Equal(t, "0", mockStream.Capture().VersionInfo)
	drs.SendCDS(&cp.Cluster{})
	assert.Equal(t, "1", mockStream.Capture().Nonce)
	assert.Equal(t, "1", mockStream.Capture().VersionInfo)
	drs.SendCDS(&cp.Cluster{})
	assert.Equal(t, "2", mockStream.Capture().Nonce)
	assert.Equal(t, "2", mockStream.Capture().VersionInfo)
	mockStream.AssertExpectations(t)
}
