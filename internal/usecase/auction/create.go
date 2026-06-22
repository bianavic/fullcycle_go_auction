package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction"
	"fullcycle-auction_go/internal/entity/bid"
	"fullcycle-auction_go/internal/internal_error"
	biduc "fullcycle-auction_go/internal/usecase/bid"
	"time"
)

type AuctionInputDTO struct {
	ProductName string           `json:"product_name" binding:"required,min=1"`
	Category    string           `json:"category" binding:"required,min=2"`
	Description string           `json:"description" binding:"required,min=10,max=200"`
	Condition   ProductCondition `json:"condition" binding:"oneof=0 1 2"`
}

type AuctionOutputDTO struct {
	ID          string           `json:"id"`
	ProductName string           `json:"product_name"`
	Category    string           `json:"category"`
	Description string           `json:"description"`
	Condition   ProductCondition `json:"condition"`
	Status      AuctionStatus    `json:"status"`
	Timestamp   time.Time        `json:"timestamp" time_format:"2006-01-02 15:04:05"`
}

type WinningInfoOutputDTO struct {
	Auction AuctionOutputDTO    `json:"auction"`
	Bid     *biduc.BidOutputDTO `json:"bid,omitempty"`
}

func New(
	auctionRepository auction.AuctionRepository,
	bidRepository bid.BidRepository,
) UseCase {
	return &useCase{
		auctionRepository: auctionRepository,
		bidRepository:     bidRepository,
	}
}

type UseCase interface {
	CreateAuction(
		ctx context.Context,
		auctionInput AuctionInputDTO) *internal_error.InternalError

	FindAuctionByID(
		ctx context.Context, id string) (*AuctionOutputDTO, *internal_error.InternalError)

	FindAuctions(
		ctx context.Context,
		status AuctionStatus,
		category, productName string) ([]AuctionOutputDTO, *internal_error.InternalError)

	FindWinningBidByAuctionID(
		ctx context.Context,
		auctionID string) (*WinningInfoOutputDTO, *internal_error.InternalError)
}

type ProductCondition int64
type AuctionStatus int64

type useCase struct {
	auctionRepository auction.AuctionRepository
	bidRepository     bid.BidRepository
}

func (uc *useCase) CreateAuction(
	ctx context.Context, auctionInput AuctionInputDTO) *internal_error.InternalError {
	newAuction, err := auction.CreateAuction(
		auctionInput.ProductName,
		auctionInput.Category,
		auctionInput.Description,
		auction.ProductCondition(auctionInput.Condition))
	if err != nil {
		return err
	}

	if err := uc.auctionRepository.CreateAuction(
		ctx, newAuction); err != nil {
		return err
	}

	return nil
}
