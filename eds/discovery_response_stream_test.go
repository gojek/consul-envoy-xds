package eds_test

import (
	"github.com/gojektech/consul-envoy-xds/eds"
	"testing"

	cp "github.com/envoyproxy/go-control-plane/api"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShouldShouldStreamDiscoveryResponse(t *testing.T) {
	mockStream := &eds.MockEDSStream{}
	mockStream.On("Send", mock.AnythingOfType("*api.DiscoveryResponse")).Return(nil)
	drs := eds.NewDiscoveryResponseStream(mockStream)
	cla := &cp.ClusterLoadAssignment{}
	drs.Send(cla)
	assert.Equal(t, "type.googleapis.com/envoy.api.v2.ClusterLoadAssignment", mockStream.Capture().TypeUrl)
	capturedResponse := mockStream.Capture()
	assert.Equal(t, 1, len(capturedResponse.Resources))
	assert.Equal(t, "type.googleapis.com/envoy.api.v2.ClusterLoadAssignment", capturedResponse.Resources[0].TypeUrl)
	claBytes, _ := proto.Marshal(cla)
	assert.Equal(t, claBytes, capturedResponse.Resources[0].Value)

	mockStream.AssertExpectations(t)
}

func TestShouldShouldStreamDiscoveryResponseWithIncrementingNonceAndVersion(t *testing.T) {
	mockStream := &eds.MockEDSStream{}
	mockStream.On("Send", mock.AnythingOfType("*api.DiscoveryResponse")).Times(3).Return(nil)
	drs := eds.NewDiscoveryResponseStream(mockStream)
	drs.Send(&cp.ClusterLoadAssignment{})
	assert.Equal(t, "0", mockStream.Capture().Nonce)
	assert.Equal(t, "0", mockStream.Capture().VersionInfo)
	drs.Send(&cp.ClusterLoadAssignment{})
	assert.Equal(t, "1", mockStream.Capture().Nonce)
	assert.Equal(t, "1", mockStream.Capture().VersionInfo)
	drs.Send(&cp.ClusterLoadAssignment{})
	assert.Equal(t, "2", mockStream.Capture().Nonce)
	assert.Equal(t, "2", mockStream.Capture().VersionInfo)
	mockStream.AssertExpectations(t)
}
