package stream

import (
	"log"
	"strconv"

	"time"

	"context"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	dis "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/gojektech/consul-envoy-xds/eventctx"
	"github.com/newrelic/go-agent"
)

type EndpointDiscoveryResponseStream interface {
	SendEDS(context.Context, []*cp.ClusterLoadAssignment) error
}

type ClusterDiscoveryResponseStream interface {
	SendCDS(context.Context, []*cp.Cluster) error
}

type RouteDiscoveryResponseStream interface {
	SendRDS(context.Context, []*cp.RouteConfiguration) error
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
func (streamer *responseStream) SendEDS(ctx context.Context, cLAList []*cp.ClusterLoadAssignment) error {
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
	streamer.send(ctx, resp)
	log.Printf("sent EDS on stream: %v", resp)
	streamer.version = time.Now().UnixNano()
	streamer.nonce = time.Now().UnixNano()
	return nil
}

//Send a Cluster on current stream
func (streamer *responseStream) SendCDS(ctx context.Context, clusters []*cp.Cluster) error {
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
	streamer.send(ctx, resp)
	log.Printf("sent CDS on stream: %v", resp)
	streamer.version = time.Now().UnixNano()
	streamer.nonce = time.Now().UnixNano()
	return nil
}

//Send a Cluster on current stream
func (streamer *responseStream) SendRDS(ctx context.Context, routeConfig []*cp.RouteConfiguration) error {
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
	streamer.send(ctx, resp)
	log.Printf("sent RDS on stream: %v", resp)
	streamer.version = time.Now().UnixNano()
	streamer.nonce = time.Now().UnixNano()
	return nil
}

func (streamer *responseStream) send(ctx context.Context, resp *cp.DiscoveryResponse) error {
	txn := eventctx.NewRelicTxn(ctx)
	seg := newrelic.StartSegment(txn, resp.GetTypeUrl())
	defer seg.End()
	return streamer.stream.Send(resp)
}

//NewDiscoveryResponseStream creates a DiscoveryResponseStream
func NewDiscoveryResponseStream(stream dis.AggregatedDiscoveryService_StreamAggregatedResourcesServer) DiscoveryResponseStream {
	return &responseStream{stream: stream, nonce: time.Now().UnixNano(), version: time.Now().UnixNano()}
}
