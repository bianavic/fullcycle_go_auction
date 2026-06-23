package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction"
	"fullcycle-auction_go/internal/apperr"
	"fullcycle-auction_go/internal/usecase/bid"
)

func (uc *useCase) FindAuctionByID(
	ctx context.Context, id string) (*OutputDTO, *apperr.InternalError) {
	found, err := uc.auctionRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &OutputDTO{
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
	category, productName string) ([]OutputDTO, *apperr.InternalError) {
	auctionEntities, err := uc.auctionRepository.FindAll(
		ctx, auction.Status(status), category, productName)
	if err != nil {
		return nil, err
	}

	var auctionOutputs []OutputDTO
	for _, value := range auctionEntities {
		auctionOutputs = append(auctionOutputs, OutputDTO{
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
	auctionID string) (*WinningInfoOutputDTO, *apperr.InternalError) {
	found, err := uc.auctionRepository.FindByID(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	auctionOutputDTO := OutputDTO{
		ID:          found.ID,
		ProductName: found.ProductName,
		Category:    found.Category,
		Description: found.Description,
		Condition:   ProductCondition(found.Condition),
		Status:      AuctionStatus(found.Status),
		Timestamp:   found.Timestamp,
	}

	winningBid, err := uc.bidRepository.FindWinningByAuctionID(ctx, found.ID)
	if err != nil {
		logger.Error("", err)
		return &WinningInfoOutputDTO{
			Auction: auctionOutputDTO,
			Bid:     nil,
		}, nil
	}

	bidOutputDTO := &bid.OutputDTO{
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
