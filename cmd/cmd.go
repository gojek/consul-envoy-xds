package main

import (
	"context"
	"io"
	"log"

	cp "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:8053", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := cp.NewAggregatedDiscoveryServiceClient(conn)
	stream, err := client.StreamAggregatedResources(context.Background())
	if err != nil {
		log.Fatalf("fail to stream: %v", err)
	}

	for {
		discoveryResponse, err := stream.Recv()
		log.Printf("received discovery request\nnonce: %s, type: %s, version: %s\n",
			discoveryResponse.Nonce,
			discoveryResponse.GetTypeUrl,
			discoveryResponse.VersionInfo)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v", err)
		}
		log.Println(discoveryResponse)
	}
}
