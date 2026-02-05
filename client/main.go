package main

import (
	"context"
	"log"
	"time"

	userv1 "grpc-go-learning/gen/go/user/v1/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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
}