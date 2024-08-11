package server

import (
	"context"
	"sync"
	pb "therealbroker/api/proto"
	bm "therealbroker/internal/broker"
	"therealbroker/pkg/broker"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedBrokerServer
	messageId int
	mu        sync.Mutex
	broker    broker.Broker
}

func NewServer() *Server {
	return &Server{messageId: 0, mu: sync.Mutex{}, broker: bm.NewModule()}
}

func (s *Server) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	s.mu.Lock()
	s.messageId++
	id := s.messageId
	s.mu.Unlock()
	msg := broker.Message{Id: id, Body: req.Body, Expiration: time.Duration(time.Duration(req.ExpirationSeconds) * time.Second)}
	id, err := s.broker.Publish(ctx, req.Subject, msg)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "broker is closed")
	}
	return &pb.PublishResponse{Id: int32(id)}, nil
}

func (s *Server) Subscribe(req *pb.SubscribeRequest, stream pb.Broker_SubscribeServer) error {
	ch, err := s.broker.Subscribe(stream.Context(), req.Subject)
	if err != nil {
		return status.Errorf(codes.Unavailable, "broker is closed")
	}

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			stream.Send(&pb.MessageResponse{Body: msg.Body})
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "subscription is cancelled")
		}
	}
}

func (s *Server) Fetch(ctx context.Context, req *pb.FetchRequest) (*pb.MessageResponse, error) {
	msg, err := s.broker.Fetch(ctx, req.Subject, int(req.Id))
	if err == broker.ErrUnavailable {
		return nil, status.Errorf(codes.Unavailable, "broker is closed")
	}
	if err == broker.ErrExpiredID {
		return nil, status.Errorf(codes.InvalidArgument, "message is expired")
	}
	if err == broker.ErrInvalidID {
		return nil, status.Errorf(codes.InvalidArgument, "message id does not exits")
	}
	return &pb.MessageResponse{Body: msg.Body}, nil
}
