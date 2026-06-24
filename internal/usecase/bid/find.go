package bid

import (
	"context"
	"fullcycle-auction_go/internal/apperr"
)

func (uc *useCase) FindBidByAuctionID(
	ctx context.Context, auctionID string) ([]OutputDTO, *apperr.InternalError) {
	bidList, err := uc.BidRepository.FindByAuctionID(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	var bidOutputList []OutputDTO
	for _, bid := range bidList {
		bidOutputList = append(bidOutputList, OutputDTO{
			ID:        bid.ID,
			UserID:    bid.UserID,
			AuctionID: bid.AuctionID,
			Amount:    bid.Amount,
			Timestamp: bid.Timestamp,
		})
	}

	return bidOutputList, nil
}

func (uc *useCase) FindWinningBidByAuctionID(
	ctx context.Context, auctionID string) (*OutputDTO, *apperr.InternalError) {
	bid, err := uc.BidRepository.FindWinningByAuctionID(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	bidOutput := &OutputDTO{
		ID:        bid.ID,
		UserID:    bid.UserID,
		AuctionID: bid.AuctionID,
		Amount:    bid.Amount,
		Timestamp: bid.Timestamp,
	}

	return bidOutput, nil
}
