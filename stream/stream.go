package stream

import (
	cp "github.com/envoyproxy/go-control-plane/api"
	"google.golang.org/grpc"
)

type DiscoveryStream interface {
	Send(*cp.DiscoveryResponse) error
	Recv() (*cp.DiscoveryRequest, error)
	grpc.ServerStream
}
