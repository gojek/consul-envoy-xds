package eds

import (
	"github.com/gojektech/consul-envoy-xds/pubsub"

	cp "github.com/envoyproxy/go-control-plane/api"
)

//EndpointSubscriptionStream is stream of stream of endpoint discovery responses
type EndpointSubscriptionStream interface {
	Stream() error
}

type endpointSubscriptionStream struct {
	stream       cp.EndpointDiscoveryService_StreamEndpointsServer
	subscription *pubsub.Subscription
}

func (es *endpointSubscriptionStream) Stream() error {
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

func NewSubscriptionStream(stream cp.EndpointDiscoveryService_StreamEndpointsServer, subscription *pubsub.Subscription) EndpointSubscriptionStream {
	return &endpointSubscriptionStream{stream: stream, subscription: subscription}
}
