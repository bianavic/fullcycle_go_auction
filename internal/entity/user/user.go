package user

import (
	"context"
	"fullcycle-auction_go/internal/internal_error"
)

type User struct {
	ID   string
	Name string
}

type UserRepository interface {
	FindUserByID(
		ctx context.Context, userID string) (*User, *internal_error.InternalError)
}
