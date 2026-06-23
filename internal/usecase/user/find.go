package user

import (
	"context"
	"fullcycle-auction_go/internal/entity/user"
	"fullcycle-auction_go/internal/apperr"
)

func New(userRepository user.Repository) UseCase {
	return &useCase{
		userRepository,
	}
}

type useCase struct {
	UserRepository user.Repository
}

type OutputDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UseCase interface {
	FindUserByID(
		ctx context.Context,
		id string) (*OutputDTO, *apperr.InternalError)
}

func (uc *useCase) FindUserByID(
	ctx context.Context, id string) (*OutputDTO, *apperr.InternalError) {
	userEntity, err := uc.UserRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &OutputDTO{
		ID:   userEntity.ID,
		Name: userEntity.Name,
	}, nil
}
