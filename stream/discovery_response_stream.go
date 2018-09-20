package stream

import (
	"log"
	"strconv"

	"time"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	dis "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

type EndpointDiscoveryResponseStream interface {
	SendEDS([]*cp.ClusterLoadAssignment) error
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
	nonce   int64
	version int64
}

//Send a CLA on current stream
func (streamer *responseStream) SendEDS(cLAList []*cp.ClusterLoadAssignment) error {
	var resources []types.Any
	for _, cLA := range cLAList {
		data, err := proto.Marshal(cLA)
		if err != nil {
			return err
		}
		resources = append(resources, types.Any{
			TypeUrl: "type.googleapis.com/envoy.api.v2.ClusterLoadAssignment",
			Value:   data,
		})
	}

	resp := &cp.DiscoveryResponse{
		VersionInfo: strconv.FormatInt(int64(streamer.version), 10),
		Resources:   resources,
		TypeUrl:     "type.googleapis.com/envoy.api.v2.ClusterLoadAssignment",
		Nonce:       strconv.FormatInt(int64(streamer.nonce), 10),
	}
	streamer.stream.Send(resp)
	log.Printf("sent EDS on stream: %v", resp)
	streamer.version = time.Now().UnixNano()
	streamer.nonce = time.Now().UnixNano()
	return nil
}

//Send a Cluster on current stream
func (streamer *responseStream) SendCDS(clusters []*cp.Cluster) error {
	if len(clusters) == 0 {
		return nil
	}
	var resources []types.Any
	for _, cluster := range clusters {
		data, err := proto.Marshal(cluster)
		if err != nil {
			return err
		}
		resources = append(resources, types.Any{
			TypeUrl: "type.googleapis.com/envoy.api.v2.Cluster",
			Value:   data,
		})
	}

	resp := &cp.DiscoveryResponse{
		VersionInfo: strconv.FormatInt(int64(streamer.version), 10),
		Resources:   resources,
		TypeUrl:     "type.googleapis.com/envoy.api.v2.Cluster",
		Nonce:       strconv.FormatInt(int64(streamer.nonce), 10),
	}
	streamer.stream.Send(resp)
	log.Printf("sent CDS on stream: %v", resp)
	streamer.version = time.Now().UnixNano()
	streamer.nonce = time.Now().UnixNano()
	return nil
}

//Send a Cluster on current stream
func (streamer *responseStream) SendRDS(routeConfig []*cp.RouteConfiguration) error {
	if len(routeConfig) == 0 {
		return nil
	}
	var resources []types.Any
	for _, route := range routeConfig {
		data, err := proto.Marshal(route)
		if err != nil {
			return err
		}
		resources = append(resources, types.Any{
			TypeUrl: "type.googleapis.com/envoy.api.v2.RouteConfiguration",
			Value:   data,
		})
	}

	resp := &cp.DiscoveryResponse{
		VersionInfo: strconv.FormatInt(int64(streamer.version), 10),
		Resources:   resources,
		TypeUrl:     "type.googleapis.com/envoy.api.v2.RouteConfiguration",
		Nonce:       strconv.FormatInt(int64(streamer.nonce), 10),
	}
	streamer.stream.Send(resp)
	log.Printf("sent RDS on stream: %v", resp)
	streamer.version = time.Now().UnixNano()
	streamer.nonce = time.Now().UnixNano()
	return nil
}

//NewDiscoveryResponseStream creates a DiscoveryResponseStream
func NewDiscoveryResponseStream(stream dis.AggregatedDiscoveryService_StreamAggregatedResourcesServer) DiscoveryResponseStream {
	return &responseStream{stream: stream, nonce: time.Now().UnixNano(), version: time.Now().UnixNano()}
}
