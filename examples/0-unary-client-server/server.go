package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ghouscht/gRPC-retry-mechanisms-with-go/middleware"
	"github.com/ghouscht/gRPC-retry-mechanisms-with-go/proto/users/v1"
	"github.com/ghouscht/gRPC-retry-mechanisms-with-go/repo"
)

type Server struct {
	repo repo.Users
}

var _ users.UsersServiceServer = &Server{}

func main() {
	const (
		listen = "[::1]:8080"
	)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	server := grpc.NewServer(grpc.UnaryInterceptor(
		logging.UnaryServerInterceptor(middleware.InterceptorLogger(logger), logging.WithLogOnEvents(logging.FinishCall)),
	))
	users.RegisterUsersServiceServer(server, &Server{repo: repo.Users{}})

	lis, err := net.Listen("tcp", listen)
	if err != nil {
		panic(err)
	}

	rootCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		defer logger.Info("server stopped",
			slog.String("bin", os.Args[0]),
		)
		logger.Info("server started",
			slog.String("listen", listen),
			slog.String("bin", os.Args[0]),
		)

		if err := server.Serve(lis); err != nil {
			panic(err)
		}
	}()

	<-rootCtx.Done()
	logger.Info("server stopping", slog.String("bin", os.Args[0]))
	server.GracefulStop() // blocks until server is stopped
}

func (s *Server) GetUser(_ context.Context, req *users.GetUserRequest) (*users.GetUserResponse, error) {
	user, err := s.repo.GetUser(req.Id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &users.GetUserResponse{User: user}, nil
}

func (s *Server) GetAllUsers(_ *users.GetAllUsersRequest, _ users.UsersService_GetAllUsersServer) error {
	return status.Error(codes.Unimplemented, "not implemented")
}
