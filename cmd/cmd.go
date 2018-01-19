package main

import (
	"context"
	"io"
	"log"

	cp "github.com/envoyproxy/go-control-plane/api"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:8053", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := cp.NewEndpointDiscoveryServiceClient(conn)
	stream, err := client.StreamEndpoints(context.Background())
	if err != nil {
		log.Fatalf("fail to stream: %v", err)
	}

	for {
		discoveryResponse, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v", err)
		}
		log.Println(discoveryResponse)
	}
}
