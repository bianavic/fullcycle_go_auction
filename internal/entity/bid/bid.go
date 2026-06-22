package bid

import (
	"context"
	"fullcycle-auction_go/internal/internal_error"
	"time"

	"github.com/google/uuid"
)

type Bid struct {
	ID        string
	UserID    string
	AuctionID string
	Amount    float64
	Timestamp time.Time
}

func CreateBid(userID, auctionID string, amount float64) (*Bid, *internal_error.InternalError) {
	bid := &Bid{
		ID:        uuid.New().String(),
		UserID:    userID,
		AuctionID: auctionID,
		Amount:    amount,
		Timestamp: time.Now(),
	}

	if err := bid.Validate(); err != nil {
		return nil, err
	}

	return bid, nil
}

func (b *Bid) Validate() *internal_error.InternalError {
	if err := uuid.Validate(b.UserID); err != nil {
		return internal_error.NewBadRequestError("UserID is not a valid id")
	} else if err := uuid.Validate(b.AuctionID); err != nil {
		return internal_error.NewBadRequestError("AuctionID is not a valid id")
	} else if b.Amount <= 0 {
		return internal_error.NewBadRequestError("Amount is not a valid value")
	}

	return nil
}

type BidRepository interface {
	CreateBid(
		ctx context.Context,
		bidEntities []Bid) *internal_error.InternalError

	FindBidByAuctionID(
		ctx context.Context, auctionID string) ([]Bid, *internal_error.InternalError)

	FindWinningBidByAuctionID(
		ctx context.Context, auctionID string) (*Bid, *internal_error.InternalError)
}
