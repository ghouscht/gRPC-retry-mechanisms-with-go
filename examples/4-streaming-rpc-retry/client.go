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

	userId, err := parseArguments(os.Args)
	if err != nil {
		panic(err)
	}

	connectCtx, cancel := context.WithTimeout(rootCtx, 3*time.Second)
	defer cancel()

	// Create a new connection to the users service server.
	clientConn, err := connect(connectCtx, "localhost:8080")
	if err != nil {
		panic(err)
	}
	defer clientConn.Close()

	// Create a new users service client stub with the connection we just created.
	usersService := users.NewUsersServiceClient(clientConn)

	var (
		backoff = elastic.NewExponentialBackoff(500*time.Millisecond, time.Second*5)
	)

	reconnect := func() (users.UsersService_GetAllUsersClient, error) {
		for retry := 0; ; retry++ {
			// Call the GetAllUsers stream RPC method with the user ID we got from the command line.
			usersClient, err := usersService.GetAllUsers(rootCtx, &users.GetAllUsersRequest{
				Offset: userId, // Start from the user ID we got from the command line or from the last user ID we received.
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
		// Read from the stream until the server closes it (EOF).
		resp, err := usersClient.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			// If the server returns an UNAVAILABLE status code, we try to reconnect and restart the stream.
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

		userId++ // Increment the user ID so we can request the next user in case the stream is restarted.
	}
}

func connect(ctx context.Context, target string, dialOptions ...grpc.DialOption) (*grpc.ClientConn, error) {
	defaultDialOptions := []grpc.DialOption{
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
}`),
	}

	return grpc.DialContext(ctx, target, append(defaultDialOptions, dialOptions...)...)
}

func parseArguments(args []string) (int64, error) {
	if len(args) != 2 {
		return 0, fmt.Errorf("usage: %s ID", args[0])
	}

	return strconv.ParseInt(args[1], 10, 64)
}
