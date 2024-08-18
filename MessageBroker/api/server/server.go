package server

import (
	"context"
	"log"
	"sync"
	pb "therealbroker/api/proto"
	bm "therealbroker/internal/broker"
	datacontrol "therealbroker/internal/data_control"
	"therealbroker/pkg/broker"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedBrokerServer
	mu     sync.Mutex
	broker broker.Broker
}

func NewServer(data datacontrol.DataControl) *Server {
	return &Server{mu: sync.Mutex{}, broker: bm.NewModule(data)}
}

func (s *Server) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	msg := broker.Message{Body: req.Body, Expiration: time.Duration(time.Duration(req.ExpirationSeconds) * time.Second)}
	id, err := s.broker.Publish(ctx, req.Subject, msg)
	if err == broker.ErrAlreadyExistID {
		return nil, status.Errorf(codes.InvalidArgument, "message id already exists")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}
	return &pb.PublishResponse{Id: id}, nil
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
	msg, err := s.broker.Fetch(ctx, req.Subject, req.Id)
	if err == broker.ErrUnavailable {
		return nil, status.Errorf(codes.Unavailable, "broker is closed")
	}
	if err == broker.ErrExpiredID {
		return nil, status.Errorf(codes.InvalidArgument, "message is expired")
	}
	if err == broker.ErrInvalidID {
		return nil, status.Errorf(codes.InvalidArgument, "message id does not exits")
	}
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "internal error")
	}
	return &pb.MessageResponse{Body: msg.Body}, nil
}
