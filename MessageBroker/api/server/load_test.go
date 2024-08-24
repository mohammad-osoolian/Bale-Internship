package server

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"therealbroker/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	serverAddr = "192.168.49.2:30000" // Update this to your server address
	reqPerSec  = 30000
	duration   = 30 * time.Second
)

func TestPublishLoad(t *testing.T) {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := proto.NewBrokerClient(conn)
	numRequests := reqPerSec * int(duration.Seconds()) * 2
	delay := time.Duration(time.Second.Nanoseconds() / reqPerSec)
	log.Println(delay)
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	timeout := time.NewTicker(duration)
	defer timeout.Stop()
	results := make(chan error, numRequests)
	var wg sync.WaitGroup

	StartTest(ticker, timeout, &wg, results, client)

	wg.Wait()
	close(results)

	for err := range results {
		if err != nil {
			t.Errorf("Request failed: %v", err)
			return
		}
	}
}

func StartTest(ticker, timeout *time.Ticker, wg *sync.WaitGroup, results chan error, client proto.BrokerClient) {
	for {
		select {
		case <-ticker.C:
			wg.Add(1)
			go func() {
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
			}()
		case <-timeout.C:
			log.Println("test finished")
			return
		}
	}
}
