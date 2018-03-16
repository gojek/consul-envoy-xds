package pubsub

import (
	"testing"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/stretchr/testify/assert"
)

func TestShouldAddSubscriptionToListOfSubscribers(t *testing.T) {
	hub := NewHub()
	subscription := hub.Subscribe()
	cla := &cp.ClusterLoadAssignment{}
	cluster := &cp.Cluster{}
	event := &Event{cla, []*cp.Cluster{cluster}, nil}
	hub.Publish(event)
	a := <-subscription.Events
	assert.Equal(t, 1, hub.Size())
	assert.Equal(t, event, a)
}

func TestShouldRemoveFromListOfSubscribersOnUnsubscribe(t *testing.T) {
	hub := NewHub()
	subscription := hub.Subscribe()
	assert.Equal(t, 1, hub.Size())
	subscription.Close()
	assert.Equal(t, 0, hub.Size())
}
