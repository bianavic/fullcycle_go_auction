package bid

import (
	"context"
	"fullcycle-auction_go/internal/internal_error"
)

func (uc *useCase) FindBidByAuctionID(
	ctx context.Context, auctionID string) ([]BidOutputDTO, *internal_error.InternalError) {
	bidList, err := uc.BidRepository.FindBidByAuctionID(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	var bidOutputList []BidOutputDTO
	for _, bid := range bidList {
		bidOutputList = append(bidOutputList, BidOutputDTO{
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
	ctx context.Context, auctionID string) (*BidOutputDTO, *internal_error.InternalError) {
	bid, err := uc.BidRepository.FindWinningBidByAuctionID(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	bidOutput := &BidOutputDTO{
		ID:        bid.ID,
		UserID:    bid.UserID,
		AuctionID: bid.AuctionID,
		Amount:    bid.Amount,
		Timestamp: bid.Timestamp,
	}

	return bidOutput, nil
}
