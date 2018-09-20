package stream

import (
	"io"
	"log"

	cp "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/gojektech/consul-envoy-xds/eds"
	"github.com/gojektech/consul-envoy-xds/pubsub"
)

//SubscriptionStream is stream of stream of x discovery responses
type SubscriptionStream interface {
	Stream() error
}

type subscriptionStream struct {
	stream       cp.AggregatedDiscoveryService_StreamAggregatedResourcesServer
	subscription *pubsub.Subscription
	service      eds.Endpoint
	hub          pubsub.Hub
}

func (es *subscriptionStream) Stream() error {
	var terminate chan bool

	go func() {
		for {
			in, err := es.stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				// log.Printf("failed to receive message on stream: %v", err)
			} else if in.VersionInfo == "" {
				log.Printf("received discovery request on stream: %v", in)
				es.hub.Publish(&pubsub.Event{CLA: es.service.CLA(), Clusters: es.service.Clusters(), Routes: es.service.Routes()})
			} else {
				log.Printf("received ACK on stream: %v", in)
			}
		}
	}()

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
			log.Printf("stream context done")
			es.subscription.Close()
			terminate <- true
		}
	}()
	<-terminate
	return nil
}

func NewSubscriptionStream(stream cp.AggregatedDiscoveryService_StreamAggregatedResourcesServer, subscription *pubsub.Subscription, service eds.Endpoint, hub pubsub.Hub) SubscriptionStream {
	return &subscriptionStream{stream: stream, subscription: subscription, service: service, hub: hub}
}
