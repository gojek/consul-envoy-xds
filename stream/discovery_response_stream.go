package stream

import (
	"strconv"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	dis "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

type EndpointDiscoveryResponseStream interface {
	SendEDS(*cp.ClusterLoadAssignment) error
}

type ClusterDiscoveryResponseStream interface {
	SendCDS([]*cp.Cluster) error
}

type RouteDiscoveryResponseStream interface {
	SendRDS([]*cp.RouteConfiguration) error
}

//DiscoveryResponseStream is an xDS Stream wrapper and wraps grpc stream API and pipes DiscoveryResponse events to it.
type DiscoveryResponseStream interface {
	ClusterDiscoveryResponseStream
	EndpointDiscoveryResponseStream
	RouteDiscoveryResponseStream
}

type responseStream struct {
	stream  dis.AggregatedDiscoveryService_StreamAggregatedResourcesServer
	nonce   int
	version int
}

//Send a CLA on current stream
func (streamer *responseStream) SendEDS(c *cp.ClusterLoadAssignment) error {
	data, err := proto.Marshal(c)
	if err != nil {
		return err
	}
	resources := []types.Any{{
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
func (streamer *responseStream) SendCDS(c []*cp.Cluster) error {
	if len(c) == 0 {
		return nil
	}
	data, err := proto.Marshal(c[0])
	if err != nil {
		return err
	}
	resources := []types.Any{{
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

//Send a Cluster on current stream
func (streamer *responseStream) SendRDS(c []*cp.RouteConfiguration) error {
	if len(c) == 0 {
		return nil
	}
	data, err := proto.Marshal(c[0])
	if err != nil {
		return err
	}
	resources := []types.Any{{
		TypeUrl: "type.googleapis.com/envoy.api.v2.RouteConfiguration",
		Value:   data,
	}}

	streamer.stream.Send(&cp.DiscoveryResponse{
		VersionInfo: strconv.FormatInt(int64(streamer.version), 10),
		Resources:   resources,
		TypeUrl:     "type.googleapis.com/envoy.api.v2.RouteConfiguration",
		Nonce:       strconv.FormatInt(int64(streamer.nonce), 10),
	})
	streamer.version++
	streamer.nonce++
	return nil
}

//NewDiscoveryResponseStream creates a DiscoveryResponseStream
func NewDiscoveryResponseStream(stream dis.AggregatedDiscoveryService_StreamAggregatedResourcesServer) DiscoveryResponseStream {
	return &responseStream{stream: stream, nonce: 0, version: 0}
}
