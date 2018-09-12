package pubsub

import (
	"testing"

	"context"
	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/stretchr/testify/assert"
)

func TestShouldAddSubscriptionToListOfSubscribers(t *testing.T) {
	hub := NewHub()
	subscription := hub.Subscribe()
	cla := []*cp.ClusterLoadAssignment{}
	cluster := &cp.Cluster{}
	event := &Event{Context: context.Background(), CLA: cla, Clusters: []*cp.Cluster{cluster}, Routes: nil}
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
