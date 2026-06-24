package bid

import (
	"context"
	"fullcycle-auction_go/internal/infra/api/web/httperr"
	"fullcycle-auction_go/internal/infra/api/web/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (u *Controller) FindBidByAuctionID(c *gin.Context) {
	auctionID := c.Param("auctionId")

	if errRest := validation.ValidateUUID(auctionID, "auctionId"); errRest != nil {
		c.JSON(errRest.Code, errRest)
		return
	}

	bidOutputList, err := u.bid.FindBidByAuctionID(context.Background(), auctionID)
	if err != nil {
		errRest := httperr.ConvertError(err)
		c.JSON(errRest.Code, errRest)
		return
	}

	c.JSON(http.StatusOK, bidOutputList)
}
