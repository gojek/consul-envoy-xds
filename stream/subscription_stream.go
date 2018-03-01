package stream

import (
	"github.com/gojektech/consul-envoy-xds/pubsub"
)

//SubscriptionStream is stream of stream of x discovery responses
type SubscriptionStream interface {
	Stream() error
}

type subscriptionStream struct {
	stream       DiscoveryStream
	subscription *pubsub.Subscription
}

func (es *subscriptionStream) Stream() error {
	var terminate (chan bool)

	go func() {
		responseStream := NewDiscoveryResponseStream(es.stream)
		for {
			select {
			case cla := <-es.subscription.Cla:
				if cla != nil {
					responseStream.Send(cla)
				}
			}
		}
	}()
	go func() {
		select {
		case <-es.stream.Context().Done():
			es.subscription.Close()
			terminate <- true
		}
	}()
	<-terminate
	return nil
}

func NewSubscriptionStream(stream DiscoveryStream, subscription *pubsub.Subscription) SubscriptionStream {
	return &subscriptionStream{stream: stream, subscription: subscription}
}
