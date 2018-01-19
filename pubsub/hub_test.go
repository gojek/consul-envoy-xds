package pubsub

import (
	"testing"

	cp "github.com/envoyproxy/go-control-plane/api"
	"github.com/stretchr/testify/assert"
)

func TestShouldAddSubscriptionToListOfSubscribers(t *testing.T) {
	hub := NewHub()
	subscription := hub.Subscribe()
	assignment := &cp.ClusterLoadAssignment{}
	hub.Publish(assignment)
	a := <-subscription.Cla
	assert.Equal(t, 1, hub.Size())
	assert.Equal(t, assignment, a)
}

func TestShouldRemoveFromListOfSubscribersOnUnsubscribe(t *testing.T) {
	hub := NewHub()
	subscription := hub.Subscribe()
	assert.Equal(t, 1, hub.Size())
	subscription.Close()
	assert.Equal(t, 0, hub.Size())
}
