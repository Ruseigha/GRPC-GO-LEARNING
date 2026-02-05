package main

import (
	"context"
	"io"
	"log"
	"time"

	userv1 "grpc-go-learning/gen/go/user/v1/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)


func testServerStreaming(client userv1.UserServiceClient)  {
	log.Println("\n========== Server-Side Streaming ==========")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()


	// Call the StreamNotifications RPC
	stream, err := client.StreamNotifications(ctx, &userv1.StreamNotificationsRequest{
		UserId: "user_1",
	})

	if err != nil {
    log.Fatalf("StreamNotifications failed: %v", err)
  }

	// Receive messages in a loop
	for {
		notification, err := stream.Recv()

		// Check if stream is done
    if err == io.EOF {
      log.Println("✅ Stream closed by server (all notifications received)")
      break
    }

		// Check for other errors
    if err != nil {
      log.Fatalf("Error receiving notification: %v", err)
    }

		// Process the notification
    log.Printf("📬 Received: [%s] %s - %s",
      notification.Type,
      notification.Title,
      notification.Message,
    )
	}
}

func testClientStreaming(client userv1.UserServiceClient)  {
	log.Println("\n========== Client-Side Streaming ==========")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
  defer cancel()


	// Open the stream
  stream, err := client.UploadUserData(ctx)
  if err != nil {
    log.Fatalf("UploadUserData failed: %v", err)
  }

	// 1. Send metadata first
  err = stream.Send(&userv1.UploadUserDataRequest{
    Data: &userv1.UploadUserDataRequest_Metadata{
    	Metadata: &userv1.UserMetadata{
    	  UserId:    "user_1",
    	  Filename:  "user_data.json",
    	  TotalSize: 5000, // 5KB total
    	},
    },
  })

	if err != nil {
    log.Fatalf("Failed to send metadata: %v", err)
  }

	log.Println("📤 Sent metadata")

	// 2. Send data in 5 chunks (1KB each)
	chunkSize := 1000 // 1KB
  totalChunks := 5

	for i := 0; i < totalChunks; i++ {
		// Create fake data chunk
    data := make([]byte, chunkSize)

		for j := range data {
      data[j] = byte(i) // Fill with chunk number
    }

		// Send chunk
    err = stream.Send(&userv1.UploadUserDataRequest{
      Data: &userv1.UploadUserDataRequest_Chunk{
        Chunk: &userv1.UserDataChunk{
          Data:        data,
          ChunkNumber: int32(i),
        },
      },
    })

		if err != nil {
      log.Fatalf("Failed to send chunk %d: %v", i, err)
    }
		log.Printf("📤 Sent chunk #%d (%d bytes)", i, chunkSize)
    time.Sleep(time.Millisecond * 500) // Simulate upload time
	}

	// 3. Close the stream and receive response
  resp, err := stream.CloseAndRecv()
  if err != nil {
    log.Fatalf("Failed to receive response: %v", err)
  }
	log.Printf("✅ Upload complete! Upload ID: %s, Bytes received: %d, Success: %v", resp.UploadId, resp.BytesReceived, resp.Success)
}

func main() {
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create UserService client from connection
	client := userv1.NewUserServiceClient(conn)

	 log.Println("✅ Connected to gRPC server")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()


	// Create a new user
	createUserRe, err := client.CreateUser(ctx, &userv1.CreateUSerRequest{
		Name: "Ruffy G",
		Email: "excellenceseigha@gmail.com",
		Age: 30,
	})

	if err != nil {
		log.Fatalf("CreateUser failed: %v", err)
	}


	log.Printf("✅ User created: %+v", createUserRe.User)
  userID := createUserRe.User.UserId


	// Get the created user
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel2()

	getUserResp, err := client.GetUser(ctx2, &userv1.GetUserRequest{
		UserId: userID,
	})

	if err != nil {
		log.Fatalf("GetUser failed: %v", err)
	}

	log.Printf("✅ User retrieved: %+v", getUserResp.User)

	// TRY TO GET NON-EXISTENT USER (ERROR HANDLING)
	ctx3, cancel3 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel3()

	_, err = client.GetUser(ctx3, &userv1.GetUserRequest{
		UserId: "non-existing-user-id",
	})
	if err != nil {
		log.Printf("❌ Expected error for non-existing user: %v", err)
	}

	// Test server-side streaming
  testServerStreaming(client)
	// Test client-side streaming
  testClientStreaming(client)
}