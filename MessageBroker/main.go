package main

import (
	"log"
	"net"
	"therealbroker/api/metrics"
	pb "therealbroker/api/proto"
	"therealbroker/api/server"

	"google.golang.org/grpc"
)

// Main requirements:
//  1. All tests should be passed
//  2. Your logs should be accessible in Graylog
//  3. Basic prometheus metrics ( latency, throughput, etc. ) should be implemented
//     for every base functionality ( publish, subscribe etc. )
func main() {
	brokerServer := server.NewServer()
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(server.UnaryMetricsInterceptor()),
		grpc.StreamInterceptor(server.StreamMetricsInterceptor()),
	)
	pb.RegisterBrokerServer(grpcServer, brokerServer)

	metrics.StartMetricsServer()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("Server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
