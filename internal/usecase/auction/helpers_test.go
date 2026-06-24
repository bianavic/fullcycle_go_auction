package auction_test

import (
	"fullcycle-auction_go/internal/apperr"
	"fullcycle-auction_go/internal/entity/auction"
	"fullcycle-auction_go/internal/entity/bid"
)

type fakeAuctionRepo struct {
	created   []*auction.Auction
	createErr *apperr.InternalError

	byID    *auction.Auction
	byIDErr *apperr.InternalError

	list    []auction.Auction
	listErr *apperr.InternalError
}

type fakeBidRepo struct {
	winning    *bid.Bid
	winningErr *apperr.InternalError
}

