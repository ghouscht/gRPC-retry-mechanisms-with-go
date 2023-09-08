package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ghouscht/gRPC-retry-mechanisms-with-go/proto/users/v1"
)

func main() {
	rootCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s ID\n", os.Args[0])
		os.Exit(1)
	}

	connectCtx, cancel := context.WithTimeout(rootCtx, 3*time.Second)
	defer cancel()

	clientConn, err := grpc.DialContext(connectCtx, "localhost:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithDefaultServiceConfig(`
{
	"methodConfig": [
		{
			"name": [
				{ "service": "proto.users.v1.UsersService" }
			],
			"retryPolicy": {
				"maxAttempts": 5,
				"initialBackoff" : "0.2s",
				"maxBackoff": "5s",
				"backoffMultiplier": 3,
				"retryableStatusCodes": [ "UNAVAILABLE" ]
			}
		}
	]
}
`),
	)
	if err != nil {
		panic(err)
	}
	defer clientConn.Close()

	usersService := users.NewUsersServiceClient(clientConn)

	requestCtx, cancel := context.WithTimeout(rootCtx, 5*time.Second)
	defer cancel()

	userId, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		panic(err)
	}

	resp, err := usersService.GetUser(requestCtx, &users.GetUserRequest{Id: userId})
	if err != nil {
		panic(err)
	}

	fmt.Printf("The user with ID %d is: %s %s born on %s\n",
		resp.User.Id,
		resp.User.FirstName,
		resp.User.LastName,
		resp.User.Birthdate.AsTime().Format("Monday, 02 Jannuary 2006"),
	)
}
