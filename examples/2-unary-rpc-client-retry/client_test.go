package main

import (
	"context"
	"net"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestConnect(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)

	go func() {
		s := grpc.NewServer()

		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	cc, err := connect(context.Background(), "localhost:8080", grpc.WithContextDialer(bufDialer))
	if err != nil {
		t.Fatal(err)
	}

	mc := cc.GetMethodConfig("/proto.users.v1.UsersService/GetUser")
	if mc.RetryPolicy == nil {
		t.Fatal("expected retry policy to be set")
	}

	spew.Dump(mc)
}
