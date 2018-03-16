package stream

import (
	cp "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/gojektech/consul-envoy-xds/pubsub"
)

//SubscriptionStream is stream of stream of x discovery responses
type SubscriptionStream interface {
	Stream() error
}

type subscriptionStream struct {
	stream       cp.AggregatedDiscoveryService_StreamAggregatedResourcesServer
	subscription *pubsub.Subscription
}

func (es *subscriptionStream) Stream() error {
	var terminate (chan bool)

	go func() {
		responseStream := NewDiscoveryResponseStream(es.stream)
		for {
			select {
			case e := <-es.subscription.Events:
				if e != nil {
					responseStream.SendCDS(e.Clusters)
					responseStream.SendRDS(e.Routes)
					responseStream.SendEDS(e.CLA)
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

func NewSubscriptionStream(stream cp.AggregatedDiscoveryService_StreamAggregatedResourcesServer, subscription *pubsub.Subscription) SubscriptionStream {
	return &subscriptionStream{stream: stream, subscription: subscription}
}
