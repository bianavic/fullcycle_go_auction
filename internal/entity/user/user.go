package user

import (
	"context"
	"fullcycle-auction_go/internal/apperr"
)

type User struct {
	ID   string
	Name string
}

type Repository interface {
	FindByID(
		ctx context.Context, userID string) (*User, *apperr.InternalError)
}
