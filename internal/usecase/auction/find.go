package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction"
	"fullcycle-auction_go/internal/internal_error"
	"fullcycle-auction_go/internal/usecase/bid"
)

func (uc *useCase) FindAuctionByID(
	ctx context.Context, id string) (*AuctionOutputDTO, *internal_error.InternalError) {
	found, err := uc.auctionRepository.FindAuctionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &AuctionOutputDTO{
		ID:          found.ID,
		ProductName: found.ProductName,
		Category:    found.Category,
		Description: found.Description,
		Condition:   ProductCondition(found.Condition),
		Status:      AuctionStatus(found.Status),
		Timestamp:   found.Timestamp,
	}, nil
}

func (uc *useCase) FindAuctions(
	ctx context.Context,
	status AuctionStatus,
	category, productName string) ([]AuctionOutputDTO, *internal_error.InternalError) {
	auctionEntities, err := uc.auctionRepository.FindAuctions(
		ctx, auction.AuctionStatus(status), category, productName)
	if err != nil {
		return nil, err
	}

	var auctionOutputs []AuctionOutputDTO
	for _, value := range auctionEntities {
		auctionOutputs = append(auctionOutputs, AuctionOutputDTO{
			ID:          value.ID,
			ProductName: value.ProductName,
			Category:    value.Category,
			Description: value.Description,
			Condition:   ProductCondition(value.Condition),
			Status:      AuctionStatus(value.Status),
			Timestamp:   value.Timestamp,
		})
	}

	return auctionOutputs, nil
}

func (uc *useCase) FindWinningBidByAuctionID(
	ctx context.Context,
	auctionID string) (*WinningInfoOutputDTO, *internal_error.InternalError) {
	found, err := uc.auctionRepository.FindAuctionByID(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	auctionOutputDTO := AuctionOutputDTO{
		ID:          found.ID,
		ProductName: found.ProductName,
		Category:    found.Category,
		Description: found.Description,
		Condition:   ProductCondition(found.Condition),
		Status:      AuctionStatus(found.Status),
		Timestamp:   found.Timestamp,
	}

	winningBid, err := uc.bidRepository.FindWinningBidByAuctionID(ctx, found.ID)
	if err != nil {
		logger.Error("", err)
		return &WinningInfoOutputDTO{
			Auction: auctionOutputDTO,
			Bid:     nil,
		}, nil
	}

	bidOutputDTO := &bid.BidOutputDTO{
		ID:        winningBid.ID,
		UserID:    winningBid.UserID,
		AuctionID: winningBid.AuctionID,
		Amount:    winningBid.Amount,
		Timestamp: winningBid.Timestamp,
	}

	return &WinningInfoOutputDTO{
		Auction: auctionOutputDTO,
		Bid:     bidOutputDTO,
	}, nil
}
