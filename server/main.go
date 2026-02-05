package main

import (
	"context"
	"fmt"
	userv1 "grpc-go-learning/gen/go/user/v1/user"
	"io"
	"log"
	"net"
	"time"

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


func (s *server) StreamNotifications(req *userv1.StreamNotificationsRequest, stream userv1.UserService_StreamNotificationsServer) error {
	log.Printf("StreamNotifications called for user_id: %s", req.UserId)

	// Validate request
  if req.UserId == "" {
      return status.Error(codes.InvalidArgument, "user_id is required")
  }

	// Simulate sending 10 notifications
	for i:=1; i<=10; i++ {
		// Check if client has disconnected
		if stream.Context().Err() != nil {
			log.Printf("Client disconnected: %v", stream.Context().Err())
      return stream.Context().Err()
		}

		// Create a notification
    notification := &userv1.Notification{
        NotificationId: fmt.Sprintf("notif_%d", i),
        UserId:         req.UserId,
        Title:          fmt.Sprintf("Notification #%d", i),
        Message:        fmt.Sprintf("This is notification number %d for user %s", i, req.UserId),
        Type:           userv1.NotificationType_NOTIFICATION_TYPE_INFO,
        Timestamp:      time.Now().Unix(),
    }


		// Send Notification
		if err := stream.Send(notification); err != nil{
			log.Printf("Failed to send notification: %v", err)
			return status.Errorf(codes.Internal, "failed to send notification: %v", err)
		}

		log.Printf("Sent notification #%d to user %s", i, req.UserId)

		// Simulate delay between notifications (1 second)
    time.Sleep(time.Second * 1)
	}

	log.Printf("Finished streaming notifications for user %s", req.UserId)
  return nil
}


// UploadUserData implements client-side streaming
func (S *server)  UploadUserData(stream userv1.UserService_UploadUserDataServer) error {
	log.Println("UploadUserData called")

	var (
    	userID       string
    	filename     string
    	totalSize    int64
    	bytesReceived int64
    	chunkCount   int
  )

	// Receive messages from client
	for {
		req, err := stream.Recv()

		// Check if client finished sending
		if err == io.EOF {
			log.Printf("Client finished sending. Received %d bytes in %d chunks", bytesReceived, chunkCount)

			// Send final response to client
			return stream.SendAndClose(&userv1.UploadUserDataResponse{
				UploadId: fmt.Sprintf("upload_%s_%d", userID, time.Now().Unix()),
				BytesReceived: bytesReceived,
				Success:       true,
			})
		}

		// Check for errors
    if err != nil {
      log.Printf("Error receiving data: %v", err)
      return status.Errorf(codes.Internal, "failed to receive data: %v", err)
    }

		// Process the received message
		switch data := req.Data.(type) {
    case *userv1.UploadUserDataRequest_Metadata:
      // First message: metadata
      userID = data.Metadata.UserId
      filename = data.Metadata.Filename
      totalSize = data.Metadata.TotalSize
      log.Printf("Receiving upload for user %s, file: %s, size: %d bytes", userID, filename, totalSize)

    case *userv1.UploadUserDataRequest_Chunk:
      // Subsequent messages: data chunks
      chunkSize := len(data.Chunk.Data)
      bytesReceived += int64(chunkSize)
      chunkCount++
      log.Printf("Received chunk #%d: %d bytes (total: %d/%d)", 
      data.Chunk.ChunkNumber, chunkSize, bytesReceived, totalSize)

    default:
      return status.Error(codes.InvalidArgument, "unknown data type")
    }
	}
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