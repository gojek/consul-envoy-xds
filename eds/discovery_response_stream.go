package eds

import (
	"strconv"

	cp "github.com/envoyproxy/go-control-plane/api"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

//DiscoveryResponseStream is a  EDS Stream wrapper and wraps grpc stream API and pipes DiscoveryResponse events to it.
type DiscoveryResponseStream interface {
	Send(*cp.ClusterLoadAssignment) error
}

type responseStream struct {
	stream  cp.EndpointDiscoveryService_StreamEndpointsServer
	nonce   int
	version int
}

//Send a CLA on current stream
func (streamer *responseStream) Send(c *cp.ClusterLoadAssignment) error {
	data, err := proto.Marshal(c)
	if err != nil {
		return err
	}
	resources := []*types.Any{{
		TypeUrl: "type.googleapis.com/envoy.api.v2.ClusterLoadAssignment",
		Value:   data,
	}}

	streamer.stream.Send(&cp.DiscoveryResponse{
		VersionInfo: strconv.FormatInt(int64(streamer.version), 10),
		Resources:   resources,
		TypeUrl:     "type.googleapis.com/envoy.api.v2.ClusterLoadAssignment",
		Nonce:       strconv.FormatInt(int64(streamer.nonce), 10),
	})
	streamer.version++
	streamer.nonce++
	return nil
}

//NewDiscoveryResponseStream creates a DiscoveryResponseStream
func NewDiscoveryResponseStream(stream cp.EndpointDiscoveryService_StreamEndpointsServer) DiscoveryResponseStream {
	return &responseStream{stream: stream, nonce: 0, version: 0}
}
