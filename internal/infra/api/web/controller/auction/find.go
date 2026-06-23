package auction

import (
	"context"
	"fullcycle-auction_go/internal/infra/api/web/httperr"
	"fullcycle-auction_go/internal/infra/api/web/validation"
	"fullcycle-auction_go/internal/usecase/auction"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (u *Controller) FindAuctionByID(c *gin.Context) {
	auctionID := c.Param("auctionId")

	if errRest := validation.ValidateUUID(auctionID, "auctionId"); errRest != nil {
		c.JSON(errRest.Code, errRest)
		return
	}

	auctionData, err := u.auctionUseCase.FindAuctionByID(context.Background(), auctionID)
	if err != nil {
		errRest := httperr.ConvertError(err)
		c.JSON(errRest.Code, errRest)
		return
	}

	c.JSON(http.StatusOK, auctionData)
}

func (u *Controller) FindAuctions(c *gin.Context) {
	status := c.Query("status")
	category := c.Query("category")
	productName := c.Query("productName")

	statusNumber, conversionError := strconv.Atoi(status)
	if conversionError != nil {
		errRest := httperr.NewBadRequestError("Error trying to validate auction status param")
		c.JSON(errRest.Code, errRest)
		return
	}

	auctions, err := u.auctionUseCase.FindAuctions(context.Background(),
		auction.AuctionStatus(statusNumber), category, productName)
	if err != nil {
		errRest := httperr.ConvertError(err)
		c.JSON(errRest.Code, errRest)
		return
	}

	c.JSON(http.StatusOK, auctions)
}

func (u *Controller) FindWinningBidByAuctionID(c *gin.Context) {
	auctionID := c.Param("auctionId")

	if errRest := validation.ValidateUUID(auctionID, "auctionId"); errRest != nil {
		c.JSON(errRest.Code, errRest)
		return
	}

	auctionData, err := u.auctionUseCase.FindWinningBidByAuctionID(context.Background(), auctionID)
	if err != nil {
		errRest := httperr.ConvertError(err)
		c.JSON(errRest.Code, errRest)
		return
	}

	c.JSON(http.StatusOK, auctionData)
}
