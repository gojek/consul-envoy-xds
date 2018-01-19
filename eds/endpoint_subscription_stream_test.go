package eds_test

import (
	"github.com/gojektech/consul-envoy-xds/eds"
	"github.com/gojektech/consul-envoy-xds/pubsub"
	"testing"
	"time"

	cp "github.com/envoyproxy/go-control-plane/api"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

func TestShouldKeepStreamingUntilInterrupted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mockStream := &eds.MockEDSStream{Ctx: ctx}

	claChan := make(pubsub.CLAChan, 1000)
	subscription := &pubsub.Subscription{ID: uuid.NewV4(), Cla: claChan, OnClose: func(subID uuid.UUID) {}}

	subscriptionStream := eds.NewSubscriptionStream(mockStream, subscription)
	done := make(chan bool, 42)
	mockStream.On("Send", mock.AnythingOfType("*api.DiscoveryResponse")).Times(42).Run(func(mock.Arguments) {
		done <- true
	}).Return(nil)

	numberOfReplies := 42
	for i := 1; i <= numberOfReplies; i++ {
		subscription.Accept(&cp.ClusterLoadAssignment{})
	}
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()
	go subscriptionStream.Stream()
	for i := 1; i <= numberOfReplies; i++ {
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
	mockStream := &eds.MockEDSStream{Ctx: ctx}

	claChan := make(pubsub.CLAChan, 1000)
	onCloseCalled := false

	subscription := &pubsub.Subscription{ID: uuid.NewV4(), Cla: claChan, OnClose: func(subID uuid.UUID) {
		onCloseCalled = true
	}}

	subscriptionStream := eds.NewSubscriptionStream(mockStream, subscription)
	go subscriptionStream.Stream()
	cancel()

	_, channelOpen := (<-subscription.Cla)
	assert.False(t, channelOpen)
	assert.True(t, onCloseCalled)
}
