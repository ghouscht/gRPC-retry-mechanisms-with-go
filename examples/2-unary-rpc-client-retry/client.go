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

	requestCtx, cancel := context.WithTimeout(rootCtx, 5*time.Second)
	defer cancel()

	// Call the GetUser RPC method with the user ID we got from the command line.
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

func connect(ctx context.Context, target string, dialOptions ...grpc.DialOption) (*grpc.ClientConn, error) {
	defaultDialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()), // disable TLS
		grpc.WithBlock(), // block until the underlying connection is up
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
}`), // Configure a retry policy for the UsersService service with a maximum of 5 attempts in case the server returns
		// an UNAVAILABLE status code.
	}

	return grpc.DialContext(ctx, target, append(defaultDialOptions, dialOptions...)...)
}

func parseArguments(args []string) (int64, error) {
	if len(args) != 2 {
		return 0, fmt.Errorf("usage: %s ID", args[0])
	}

	return strconv.ParseInt(args[1], 10, 64)
}
