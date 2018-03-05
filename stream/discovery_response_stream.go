package stream

import (
	"strconv"

	cp "github.com/envoyproxy/go-control-plane/api"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

type EndpointDiscoveryResponseStream interface {
	SendEDS(*cp.ClusterLoadAssignment) error
}

type ClusterDiscoveryResponseStream interface {
	SendCDS(*cp.Cluster) error
}

//DiscoveryResponseStream is an xDS Stream wrapper and wraps grpc stream API and pipes DiscoveryResponse events to it.
type DiscoveryResponseStream interface {
	ClusterDiscoveryResponseStream
	EndpointDiscoveryResponseStream
}

type responseStream struct {
	stream  cp.AggregatedDiscoveryService_StreamAggregatedResourcesServer
	nonce   int
	version int
}

//Send a CLA on current stream
func (streamer *responseStream) SendEDS(c *cp.ClusterLoadAssignment) error {
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

//Send a Cluster on current stream
func (streamer *responseStream) SendCDS(c *cp.Cluster) error {
	data, err := proto.Marshal(c)
	if err != nil {
		return err
	}
	resources := []*types.Any{{
		TypeUrl: "type.googleapis.com/envoy.api.v2.Cluster",
		Value:   data,
	}}

	streamer.stream.Send(&cp.DiscoveryResponse{
		VersionInfo: strconv.FormatInt(int64(streamer.version), 10),
		Resources:   resources,
		TypeUrl:     "type.googleapis.com/envoy.api.v2.Cluster",
		Nonce:       strconv.FormatInt(int64(streamer.nonce), 10),
	})
	streamer.version++
	streamer.nonce++
	return nil
}

//NewDiscoveryResponseStream creates a DiscoveryResponseStream
func NewDiscoveryResponseStream(stream cp.AggregatedDiscoveryService_StreamAggregatedResourcesServer) DiscoveryResponseStream {
	return &responseStream{stream: stream, nonce: 0, version: 0}
}
