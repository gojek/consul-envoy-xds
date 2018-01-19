package pubsub_test

import (
	"github.com/gojektech/consul-envoy-xds/pubsub"
	"testing"

	cp "github.com/envoyproxy/go-control-plane/api"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestShouldSendCLAOnCLAChanOnAccept(t *testing.T) {
	claChan := make(pubsub.CLAChan, 1000)
	subscription := &pubsub.Subscription{ID: uuid.NewV4(), Cla: claChan, OnClose: func(subID uuid.UUID) {
	}}
	publishedCla := &cp.ClusterLoadAssignment{}
	subscription.Accept(publishedCla)
	cla := <-subscription.Cla

	assert.Equal(t, cla, publishedCla)
}

func TestShouldCallOnCloseAndCloseChanWhenSubsciptionIsClosed(t *testing.T) {
	claChan := make(pubsub.CLAChan, 1000)
	onCloseCalled := false
	subscription := &pubsub.Subscription{ID: uuid.NewV4(), Cla: claChan, OnClose: func(subID uuid.UUID) {
		onCloseCalled = true
	}}
	subscription.Close()

	_, channelOpen := (<-claChan)
	assert.False(t, channelOpen)
	assert.True(t, onCloseCalled)
}
