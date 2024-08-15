package server

import (
	"context"
	"sync"
	"testing"
	"time"

	"therealbroker/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	serverAddr      = "localhost:50051" // Update this to your server address
	numRequests     = 100000
	requestInterval = 100 * time.Microsecond
)

func TestPublishLoad(t *testing.T) {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := proto.NewBrokerClient(conn)

	results := make(chan error, numRequests)
	var wg sync.WaitGroup

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(reqID int) {
			defer wg.Done()
			req := &proto.PublishRequest{
				Subject:           "Test Subject",
				Body:              "Test Body",
				ExpirationSeconds: 3600,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
			defer cancel()

			_, err := client.Publish(ctx, req)
			if err != nil {
				results <- err
				return
			}

			results <- nil
		}(i)

		// time.Sleep(requestInterval)
	}

	wg.Wait()
	close(results)

	var numErrors int
	for err := range results {
		if err != nil {
			t.Errorf("Request failed: %v", err)
			numErrors++
		}
	}

	if numErrors > 0 {
		t.Fatalf("%d requests failed", numErrors)
	}
}
