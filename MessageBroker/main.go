package main

import (
	"context"
	"log"
	"net"
	"therealbroker/api/metrics"
	"therealbroker/api/server"
	datacontrol "therealbroker/internal/data_control"

	pb "therealbroker/api/proto"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

// Main requirements:
//  1. All tests should be passed
//  2. Your logs should be accessible in Graylog
//  3. Basic prometheus metrics ( latency, throughput, etc. ) should be implemented
//     for every base functionality ( publish, subscribe etc. )
func main() {
	postgres := datacontrol.NewDataPostgres("localhost", "5432", "postgres", "8764", "TestDB", context.Background())
	postgres.Connect()
	defer postgres.Close()

	brokerServer := server.NewServer(postgres)
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
