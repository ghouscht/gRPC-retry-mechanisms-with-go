package repo

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ghouscht/gRPC-retry-mechanisms-with-go/proto/users/v1"
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

type Users struct{}

var allUsers = []*users.User{
	{
		Id:        0,
		FirstName: "Alfred",
		LastName:  "Einstein",
		Birthdate: timestamppb.New(time.Date(1879, 3, 14, 0, 0, 0, 0, time.UTC)),
	},
	{
		Id:        1,
		FirstName: "Marie",
		LastName:  "Curie",
		Birthdate: timestamppb.New(time.Date(1867, 11, 7, 0, 0, 0, 0, time.UTC)),
	},
	{
		Id:        2,
		FirstName: "Isaac",
		LastName:  "Newton",
		Birthdate: timestamppb.New(time.Date(1643, 1, 4, 0, 0, 0, 0, time.UTC)),
	},
	{
		Id:        3,
		FirstName: "Nikola",
		LastName:  "Tesla",
		Birthdate: timestamppb.New(time.Date(1856, 7, 10, 0, 0, 0, 0, time.UTC)),
	},
	{
		Id:        4,
		FirstName: "Katherine",
		LastName:  "Johnson",
		Birthdate: timestamppb.New(time.Date(1918, 8, 26, 0, 0, 0, 0, time.UTC)),
	},
}

func (u Users) GetUser(id int64) (*users.User, error) {
	for _, user := range allUsers {
		if user.Id == id {
			return user, nil
		}
	}

	return nil, ErrNotFound
}

func (u Users) GetAllUsers(offset int64) ([]*users.User, error) {
	if offset < 0 || offset >= int64(len(allUsers)) {
		return nil, ErrNotFound
	}

	return allUsers[offset:], nil
}
