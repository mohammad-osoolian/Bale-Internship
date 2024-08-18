package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"therealbroker/api/metrics"
	"therealbroker/api/server"
	"therealbroker/config"
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
	if err := config.LoadConfig(); err != nil {
		log.Println("failed to load configs")
		return
	}
	log.Println("*** config loaded ***")

	var DB datacontrol.DataControl
	if config.DATA_CONTROL == "memory" {
		memory := datacontrol.NewDataMemory()
		DB = memory
	}

	if config.DATA_CONTROL == "postgres" {
		postgres := datacontrol.NewDataPostgres(
			config.POSTGRES_HOST,
			config.POSTGRES_PORT,
			config.POSTGRES_USER,
			config.POSTGRES_PASS,
			config.POSTGRES_DBNAME,
			context.Background())

		err := postgres.Connect()
		if err != nil {
			log.Println(err)
			return
		}
		defer postgres.Close()
		DB = postgres

	}

	if config.DATA_CONTROL == "scylla" {
		scylla := datacontrol.NewDataScylla(
			config.SCYLLA_HOST,
			config.SCYLLA_PORT,
			config.SCYLLA_KEYSPACE,
			config.SCYLLA_FORGET)

		err := scylla.Connect()
		if err != nil {
			log.Println(err)
			return
		}
		defer scylla.Close()
		DB = scylla
	}

	log.Println("*** data control started on", config.DATA_CONTROL, "***")

	brokerServer := server.NewServer(DB)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(server.UnaryMetricsInterceptor()),
		grpc.StreamInterceptor(server.StreamMetricsInterceptor()),
	)
	pb.RegisterBrokerServer(grpcServer, brokerServer)

	metrics.StartMetricsServer()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GRPC_PORT))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("Server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
