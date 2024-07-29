package server

import (
	"context"
	"log"
	pb "messenger"
	"sort"
	"sync"
	"time"
)

type Server struct {
	pb.UnimplementedMessengerServer
	mu            sync.Mutex
	users         map[int32]*User
	messages      map[int32]*Message
	userID        int32
	msgID         int32
	fileServerURL string
}

type User struct {
	ID       int32
	Username string
	FileID   string
}

type Message struct {
	ID        int32
	Content   string
	Sender    int32
	Receiver  int32
	Timestamp int64
}

func NewServer() *Server {
	return &Server{
		users:         make(map[int32]*User),
		messages:      make(map[int32]*Message),
		userID:        1,
		msgID:         1,
		fileServerURL: "http://fileserver:8000",
	}
}

func (s *Server) AddUser(ctx context.Context, req *pb.AddUserRequest) (*pb.AddUserResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !IsValidUsername(req.Username) {
		return &pb.AddUserResponse{Error: "invalid username"}, nil
	}

	if !IsUniqueUsername(s, req.Username) {
		return &pb.AddUserResponse{Error: "username already exists"}, nil
	}

	valid, err := IsValidFile(s, req.FileId)
	if err != nil {
		return &pb.AddUserResponse{Error: "faild to validate file"}, nil
	}

	if !valid {
		return &pb.AddUserResponse{Error: "invalid file"}, nil
	}

	user := &User{
		ID:       s.userID,
		Username: req.Username,
		FileID:   req.FileId,
	}

	s.users[s.userID] = user
	s.userID++

	return &pb.AddUserResponse{UserId: user.ID}, nil
}

func (s *Server) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Println(req)
	sender, err := GetUserByIDOrUsername(s, req.Sender)
	if err != nil {
		log.Println(err)
		return &pb.SendMessageResponse{Error: "sender user does not exist"}, nil
	}

	receiver, err := GetUserByIDOrUsername(s, req.Receiver)
	if err != nil {
		return &pb.SendMessageResponse{Error: "receiver user does not exist"}, nil
	}

	// Check if content is valid based on type
	content, err := ValidateContent(s, req.Content)
	if err != nil {
		return &pb.SendMessageResponse{Error: "invalid content"}, nil
	}

	message := &Message{
		ID:        s.msgID,
		Content:   content,
		Sender:    sender.ID,
		Receiver:  receiver.ID,
		Timestamp: time.Now().Unix(),
	}

	s.messages[s.msgID] = message
	s.msgID++

	return &pb.SendMessageResponse{MessageId: message.ID}, nil
}

func (s *Server) FetchMessage(ctx context.Context, req *pb.FetchMessageRequest) (*pb.FetchMessageResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	message, exists := s.messages[req.MessageId]
	if !exists {
		return &pb.FetchMessageResponse{Error: "message does not exist"}, nil
	}

	return &pb.FetchMessageResponse{Content: message.Content}, nil
}

func (s *Server) GetUserMessages(ctx context.Context, req *pb.GetUserMessagesRequest) (*pb.GetUserMessagesResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[req.UserId]
	if !exists {
		return &pb.GetUserMessagesResponse{Error: "user does not exist"}, nil
	}

	var userMessages []*pb.Message
	for _, msg := range s.messages {
		if msg.Sender == user.ID || msg.Receiver == user.ID {
			userMessages = append(userMessages, &pb.Message{
				Id:        msg.ID,
				Content:   msg.Content,
				Sender:    msg.Sender,
				Receiver:  msg.Receiver,
				Timestamp: msg.Timestamp,
			})
		}
	}

	// Sort messages by timestamp (latest first)
	sort.Slice(userMessages, func(i, j int) bool {
		return userMessages[i].Timestamp > userMessages[j].Timestamp
	})

	return &pb.GetUserMessagesResponse{Messages: userMessages}, nil
}
