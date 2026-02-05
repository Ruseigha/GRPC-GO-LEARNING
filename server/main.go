package main

import (
	"context"
	"fmt"
	userv1 "grpc-go-learning/gen/go/user/v1/user"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//server implements UserServiceServer interface
type server struct {
	userv1.UnimplementedUserServiceServer
	users map[string]*userv1.User //in-memory strorage
}

//GetUser implement the GetUser RPC method
func (s *server) GetUser(ctx context.Context, req *userv1.GetUserRequest)(*userv1.GetUserResponse, error)  {
	log.Printf("GetUser called with user_id: %s", req.UserId)

	if req.UserId == ""{
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, exists := s.users[req.UserId]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "user with id %s notfound", req.UserId)
	}

	return &userv1.GetUserResponse{
		User: user,
	}, nil
}


//CreateUser implement the CreateUser RPC method
func (s *server) CreateUser(ctx context.Context, req *userv1.CreateUSerRequest) (*userv1.CreateUserResponse, error)  {
	log.Printf("CreateUser called with name: %s, email: %s", req.Name, req.Email)

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name id required")
	}

	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	//Generate user Id
	userID := fmt.Sprintf("user_%d", len(s.users)+1)

	//Create user
	user := &userv1.User{
		UserId: userID,
		Name: req.Name,
		Email: req.Email,
		Age: req.Age,
		Status: userv1.UserStatus_USER_STATUS_UNSPECIFIED,
	}

	//Store user
	s.users[userID] = user

	log.Printf("User created successfully: %s", userID)
	return &userv1.CreateUserResponse{
		User: user,
	}, nil
}


func main()  {
	// Create TCP listener on port 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	//Create gRPC server
	grpcServer := grpc.NewServer()

	// Create our server implementation with in-memory storage
	userServer := &server{
		users: make(map[string]*userv1.User),
	}

	// register our server with gRPC server
	userv1.RegisterUserServiceServer(grpcServer, userServer)

	log.Println("🚀 gRPC server listening on :50051")

	// start server
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}