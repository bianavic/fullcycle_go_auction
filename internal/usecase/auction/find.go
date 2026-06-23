package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/apperr"
	"fullcycle-auction_go/internal/entity/auction"
	"fullcycle-auction_go/internal/usecase/bid"
)

func toOutputDTO(a *auction.Auction) OutputDTO {
	return OutputDTO{
		ID:          a.ID,
		ProductName: a.ProductName,
		Category:    a.Category,
		Description: a.Description,
		Condition:   ProductCondition(a.Condition),
		Status:      AuctionStatus(a.Status),
		Timestamp:   a.Timestamp,
	}
}

func (uc *useCase) FindAuctionByID(
	ctx context.Context, id string) (*OutputDTO, *apperr.InternalError) {
	found, err := uc.auctionRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	dto := toOutputDTO(found)
	return &dto, nil
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
		auctionOutputs = append(auctionOutputs, toOutputDTO(&value))
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

	auctionOutputDTO := toOutputDTO(found)

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
