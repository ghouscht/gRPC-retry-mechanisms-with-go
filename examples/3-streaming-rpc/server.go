package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			logging.UnaryServerInterceptor(middleware.InterceptorLogger(logger), logging.WithLogOnEvents(logging.FinishCall)),
		),
		grpc.StreamInterceptor(
			logging.StreamServerInterceptor(middleware.InterceptorLogger(logger),
				logging.WithLogOnEvents(
					logging.StartCall,
					logging.PayloadSent,
					logging.FinishCall,
				),
			),
		),
	)
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
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *Server) GetAllUsers(req *users.GetAllUsersRequest, srv users.UsersService_GetAllUsersServer) error {
	allUsers, err := s.repo.GetAllUsers(req.Offset)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil
		}

		return status.Error(codes.Internal, err.Error())
	}

	for _, user := range allUsers {
		if err := srv.Send(&users.GetAllUsersResponse{User: user}); err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		if time.Now().Second()%2 == 0 {
			return status.Errorf(codes.Unavailable, "service unavailable in even seconds")
		}

		time.Sleep(250 * time.Millisecond)
	}

	return nil
}
