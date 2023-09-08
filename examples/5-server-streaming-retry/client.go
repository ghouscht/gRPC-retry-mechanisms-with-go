package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/olivere/elastic/v7"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/ghouscht/gRPC-retry-mechanisms-with-go/proto/users/v1"
)

func main() {
	rootCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s NAME\n", os.Args[0])
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

	userId, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		panic(err)
	}

	usersService := users.NewUsersServiceClient(clientConn)

	var (
		backoff = elastic.NewExponentialBackoff(500*time.Millisecond, time.Second*5)
	)

	reconnect := func() (users.UsersService_GetAllUsersClient, error) {
		for retry := 0; ; retry++ {
			usersClient, err := usersService.GetAllUsers(rootCtx, &users.GetAllUsersRequest{
				Offset: userId,
			})
			if err != nil {
				waitTime, ok := backoff.Next(retry)
				if !ok {
					return nil, fmt.Errorf("calling GetAllUsers failed, giving up after %d retries: %w", retry, err)
				}

				time.Sleep(waitTime)
				continue
			}

			return usersClient, nil
		}
	}

	usersClient, err := reconnect()
	if err != nil {
		panic(err)
	}

	for {
		resp, err := usersClient.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			if status.Code(err) == codes.Unavailable {
				usersClient, err = reconnect()
				if err != nil {
					panic(err)
				}

				continue
			}

			panic(err)
		}

		fmt.Printf("The user with ID %d is: %s %s born on %s\n",
			resp.User.Id,
			resp.User.FirstName,
			resp.User.LastName,
			resp.User.Birthdate.AsTime().Format("Monday, 02 Jannuary 2006"),
		)

		userId++
	}
}
