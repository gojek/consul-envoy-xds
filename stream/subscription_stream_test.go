package stream_test

import (
	"testing"
	"time"

	"github.com/gojek/consul-envoy-xds/pubsub"
	"github.com/gojek/consul-envoy-xds/stream"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/gojek/consul-envoy-xds/eds"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

func TestShouldKeepStreamingUntilInterrupted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mockStream := &stream.MockXDSStream{Ctx: ctx}

	eventsChan := make(pubsub.EventChan, 1000)
	subscription := &pubsub.Subscription{ID: uuid.NewV4(), Events: eventsChan, OnClose: func(subID uuid.UUID) {}}

	mockEndpoint := &eds.MockEndpoint{}
	mockEndpoint.On("CLA").Return([]*cp.ClusterLoadAssignment{})
	mockEndpoint.On("Clusters").Return([]*cp.Cluster{})
	mockEndpoint.On("Routes").Return([]*cp.RouteConfiguration{})
	hub := pubsub.NewHub()
	subscriptionStream := stream.NewSubscriptionStream(mockStream, subscription, mockEndpoint, hub)
	n := 42
	done := make(chan bool, n)
	mockStream.On("Recv").Return(&cp.DiscoveryRequest{}, nil)
	mockStream.On("Send", mock.AnythingOfType("*v2.DiscoveryResponse")).Times(n * 3).Run(func(mock.Arguments) {
		done <- true
	}).Return(nil)

	for i := 1; i <= n; i++ {
		subscription.Accept(&pubsub.Event{CLA: []*cp.ClusterLoadAssignment{}, Clusters: []*cp.Cluster{{}}, Routes: []*cp.RouteConfiguration{{}}})
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

func TestShouldRespondToRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mockStream := &stream.MockXDSStream{Ctx: ctx}

	hub := pubsub.NewHub()
	subscription := hub.Subscribe()

	mockEndpoint := &eds.MockEndpoint{}
	mockEndpoint.On("CLA").Return([]*cp.ClusterLoadAssignment{})
	mockEndpoint.On("Clusters").Return([]*cp.Cluster{})
	mockEndpoint.On("Routes").Return([]*cp.RouteConfiguration{})
	subscriptionStream := stream.NewSubscriptionStream(mockStream, subscription, mockEndpoint, hub)
	done := make(chan bool, 1)
	mockStream.On("Recv").Return(&cp.DiscoveryRequest{}, nil)
	mockStream.On("Send", mock.AnythingOfType("*v2.DiscoveryResponse")).Run(func(mock.Arguments) {
		done <- true
	}).Return(nil)

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()

	go subscriptionStream.Stream()

	for i := 0; i < 3; i++ {
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
	mockEndpoint.AssertExpectations(t)
}

func TestShouldNotRespondToACK(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mockStream := &stream.MockXDSStream{Ctx: ctx}

	hub := pubsub.NewHub()
	subscription := hub.Subscribe()

	mockEndpoint := &eds.MockEndpoint{}
	subscriptionStream := stream.NewSubscriptionStream(mockStream, subscription, mockEndpoint, hub)
	done := make(chan bool, 1)
	mockStream.On("Recv").Run(func(mock.Arguments) {
		done <- true
	}).Return(&cp.DiscoveryRequest{VersionInfo: "123"}, nil)

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()

	go subscriptionStream.Stream()

	for i := 0; i < 3; i++ {
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
	mockEndpoint.AssertExpectations(t)
}

func TestShouldCloseSubscriptionOnInterrupted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mockStream := &stream.MockXDSStream{Ctx: ctx}
	mockStream.On("Recv").Return(&cp.DiscoveryRequest{}, nil)

	eventsChan := make(pubsub.EventChan, 1000)
	onCloseCalled := false

	subscription := &pubsub.Subscription{ID: uuid.NewV4(), Events: eventsChan, OnClose: func(subID uuid.UUID) {
		onCloseCalled = true
	}}

	mockEndpoint := &eds.MockEndpoint{}
	mockEndpoint.On("CLA").Return([]*cp.ClusterLoadAssignment{})
	mockEndpoint.On("Clusters").Return([]*cp.Cluster{})
	mockEndpoint.On("Routes").Return([]*cp.RouteConfiguration{})
	hub := pubsub.NewHub()
	subscriptionStream := stream.NewSubscriptionStream(mockStream, subscription, mockEndpoint, hub)
	go subscriptionStream.Stream()
	cancel()

	_, channelOpen := (<-subscription.Events)
	assert.False(t, channelOpen)
	assert.True(t, onCloseCalled)
}
