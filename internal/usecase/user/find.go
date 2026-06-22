package user

import (
	"context"
	"fullcycle-auction_go/internal/entity/user"
	"fullcycle-auction_go/internal/internal_error"
)

func New(userRepository user.UserRepository) UseCase {
	return &useCase{
		userRepository,
	}
}

type useCase struct {
	UserRepository user.UserRepository
}

type UserOutputDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UseCase interface {
	FindUserByID(
		ctx context.Context,
		id string) (*UserOutputDTO, *internal_error.InternalError)
}

func (uc *useCase) FindUserByID(
	ctx context.Context, id string) (*UserOutputDTO, *internal_error.InternalError) {
	userEntity, err := uc.UserRepository.FindUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &UserOutputDTO{
		ID:   userEntity.ID,
		Name: userEntity.Name,
	}, nil
}
