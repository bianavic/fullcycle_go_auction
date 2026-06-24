package auction

import (
	"context"
	"fullcycle-auction_go/internal/apperr"
	"fullcycle-auction_go/internal/entity/auction"
	"fullcycle-auction_go/internal/entity/bid"
	biduc "fullcycle-auction_go/internal/usecase/bid"
	"time"
)

type InputDTO struct {
	ProductName string           `json:"product_name" binding:"required,min=1"`
	Category    string           `json:"category" binding:"required,min=2"`
	Description string           `json:"description" binding:"required,min=10,max=200"`
	Condition   ProductCondition `json:"condition" binding:"oneof=0 1 2"`
}

type OutputDTO struct {
	ID          string           `json:"id"`
	ProductName string           `json:"product_name"`
	Category    string           `json:"category"`
	Description string           `json:"description"`
	Condition   ProductCondition `json:"condition"`
	Status      AuctionStatus    `json:"status"`
	Timestamp   time.Time        `json:"timestamp" time_format:"2006-01-02 15:04:05"`
}

type WinningInfoOutputDTO struct {
	Auction OutputDTO        `json:"auction"`
	Bid     *biduc.OutputDTO `json:"bid,omitempty"`
}

func New(
	auctionRepository auction.Repository,
	bidRepository bid.Repository,
) UseCase {
	return &useCase{
		auctionRepository: auctionRepository,
		bidRepository:     bidRepository,
	}
}

type UseCase interface {
	CreateAuction(
		ctx context.Context,
		auctionInput InputDTO) *apperr.InternalError

	FindAuctionByID(
		ctx context.Context, id string) (*OutputDTO, *apperr.InternalError)

	FindAuctions(
		ctx context.Context,
		status AuctionStatus,
		category, productName string) ([]OutputDTO, *apperr.InternalError)

	FindWinningBidByAuctionID(
		ctx context.Context,
		auctionID string) (*WinningInfoOutputDTO, *apperr.InternalError)
}

type ProductCondition int64
type AuctionStatus int64

type useCase struct {
	auctionRepository auction.Repository
	bidRepository     bid.Repository
}

func (uc *useCase) CreateAuction(
	ctx context.Context, auctionInput InputDTO) *apperr.InternalError {
	newAuction, err := auction.Create(
		auctionInput.ProductName,
		auctionInput.Category,
		auctionInput.Description,
		auction.ProductCondition(auctionInput.Condition))
	if err != nil {
		return err
	}

	if err := uc.auctionRepository.Create(
		ctx, newAuction); err != nil {
		return err
	}

	return nil
}
