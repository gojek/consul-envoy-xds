package stream_test

import (
	"testing"
	"time"

	"github.com/gojektech/consul-envoy-xds/pubsub"
	"github.com/gojektech/consul-envoy-xds/stream"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

func TestShouldKeepStreamingUntilInterrupted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mockStream := &stream.MockXDSStream{Ctx: ctx}

	eventsChan := make(pubsub.EventChan, 1000)
	subscription := &pubsub.Subscription{ID: uuid.NewV4(), Events: eventsChan, OnClose: func(subID uuid.UUID) {}}

	subscriptionStream := stream.NewSubscriptionStream(mockStream, subscription)
	n := 42
	done := make(chan bool, n)
	mockStream.On("Send", mock.AnythingOfType("*v2.DiscoveryResponse")).Times(n * 3).Run(func(mock.Arguments) {
		done <- true
	}).Return(nil)

	for i := 1; i <= n; i++ {
		subscription.Accept(&pubsub.Event{&cp.ClusterLoadAssignment{}, []*cp.Cluster{&cp.Cluster{}}, []*cp.RouteConfiguration{&cp.RouteConfiguration{}}})
	}
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()
	go subscriptionStream.Stream()
	for i := 1; i <= n*3; i++ {
		select {
		case <-done:
			t.Logf("%d was done\n", i)
		case <-timeout:
			cancel()
			t.Log("Failing after timeout")
			t.FailNow()
		}
	}
	cancel()
	mockStream.AssertExpectations(t)
}

func TestShouldCloseSubscriptionOnInterrupted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mockStream := &stream.MockXDSStream{Ctx: ctx}

	eventsChan := make(pubsub.EventChan, 1000)
	onCloseCalled := false

	subscription := &pubsub.Subscription{ID: uuid.NewV4(), Events: eventsChan, OnClose: func(subID uuid.UUID) {
		onCloseCalled = true
	}}

	subscriptionStream := stream.NewSubscriptionStream(mockStream, subscription)
	go subscriptionStream.Stream()
	cancel()

	_, channelOpen := (<-subscription.Events)
	assert.False(t, channelOpen)
	assert.True(t, onCloseCalled)
}
