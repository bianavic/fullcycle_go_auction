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

const (
	activeStatus    = 0
	completedStatus = 1
)

func (ctrl *Controller) FindAuctionByID(c *gin.Context) {
	auctionID := c.Param("auctionId")

	if errRest := validation.ValidateUUID(auctionID, "auctionId"); errRest != nil {
		c.JSON(errRest.Code, errRest)
		return
	}

	auctionData, err := ctrl.auctionUseCase.FindAuctionByID(context.Background(), auctionID)
	if err != nil {
		errRest := httperr.ConvertError(err)
		c.JSON(errRest.Code, errRest)
		return
	}

	c.JSON(http.StatusOK, auctionData)
}

func (ctrl *Controller) FindAuctions(c *gin.Context) {
	status := c.Query("status")
	category := c.Query("category")
	productName := c.Query("productName")

	// Sem o parâmetro status, lista leilões de qualquer status.
	var statusFilter *auction.AuctionStatus
	if status != "" {
		statusNumber, conversionError := strconv.Atoi(status)
		if conversionError != nil || (statusNumber != activeStatus && statusNumber != completedStatus) {
			errRest := httperr.NewBadRequestError("error trying to validate auction status param")
			c.JSON(errRest.Code, errRest)
			return
		}

		parsed := auction.AuctionStatus(statusNumber)
		statusFilter = &parsed
	}

	auctions, err := ctrl.auctionUseCase.FindAuctions(context.Background(),
		statusFilter, category, productName)
	if err != nil {
		errRest := httperr.ConvertError(err)
		c.JSON(errRest.Code, errRest)
		return
	}

	c.JSON(http.StatusOK, auctions)
}

func (ctrl *Controller) FindWinningBidByAuctionID(c *gin.Context) {
	auctionID := c.Param("auctionId")

	if errRest := validation.ValidateUUID(auctionID, "auctionId"); errRest != nil {
		c.JSON(errRest.Code, errRest)
		return
	}

	auctionData, err := ctrl.auctionUseCase.FindWinningBidByAuctionID(context.Background(), auctionID)
	if err != nil {
		errRest := httperr.ConvertError(err)
		c.JSON(errRest.Code, errRest)
		return
	}

	c.JSON(http.StatusOK, auctionData)
}
