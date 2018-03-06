package pubsub_test

import (
	"testing"

	"github.com/gojektech/consul-envoy-xds/pubsub"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestShouldSendCLAOnCLAChanOnAccept(t *testing.T) {
	eventsChan := make(pubsub.EventChan, 1000)
	subscription := &pubsub.Subscription{ID: uuid.NewV4(), Events: eventsChan, OnClose: func(subID uuid.UUID) {
	}}
	publishedCla := &cp.ClusterLoadAssignment{}
	cluster := &cp.Cluster{}
	event := &pubsub.Event{publishedCla, []*cp.Cluster{cluster}, []*cp.RouteConfiguration{&cp.RouteConfiguration{}}}
	subscription.Accept(event)
	e := <-subscription.Events

	assert.Equal(t, e, event)
}

func TestShouldCallOnCloseAndCloseChanWhenSubsciptionIsClosed(t *testing.T) {
	eventsChan := make(pubsub.EventChan, 1000)
	onCloseCalled := false
	subscription := &pubsub.Subscription{ID: uuid.NewV4(), Events: eventsChan, OnClose: func(subID uuid.UUID) {
		onCloseCalled = true
	}}
	subscription.Close()

	_, channelOpen := (<-eventsChan)
	assert.False(t, channelOpen)
	assert.True(t, onCloseCalled)
}
