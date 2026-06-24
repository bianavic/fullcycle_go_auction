package bid

import (
	"context"
	"fullcycle-auction_go/internal/apperr"
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

func Create(userID, auctionID string, amount float64) (*Bid, *apperr.InternalError) {
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

func (b *Bid) Validate() *apperr.InternalError {
	if err := uuid.Validate(b.UserID); err != nil {
		return apperr.NewBadRequestError("UserID is not a valid id")
	} else if err := uuid.Validate(b.AuctionID); err != nil {
		return apperr.NewBadRequestError("AuctionID is not a valid id")
	} else if b.Amount <= 0 {
		return apperr.NewBadRequestError("amount is not a valid value")
	}

	return nil
}

type Repository interface {
	Create(
		ctx context.Context,
		bidEntities []Bid) *apperr.InternalError

	FindByAuctionID(
		ctx context.Context, auctionID string) ([]Bid, *apperr.InternalError)

	FindWinningByAuctionID(
		ctx context.Context, auctionID string) (*Bid, *apperr.InternalError)
}
