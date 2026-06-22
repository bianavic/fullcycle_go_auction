package bid

import (
	"context"
	"fullcycle-auction_go/configuration/rest_err"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (u *BidController) FindBidByAuctionID(c *gin.Context) {
	auctionID := c.Param("auctionId")

	if err := uuid.Validate(auctionID); err != nil {
		errRest := rest_err.NewBadRequestError("Invalid fields", rest_err.Causes{
			Field:   "auctionId",
			Message: "Invalid UUID value",
		})

		c.JSON(errRest.Code, errRest)
		return
	}

	bidOutputList, err := u.bid.FindBidByAuctionID(context.Background(), auctionID)
	if err != nil {
		errRest := rest_err.ConvertError(err)
		c.JSON(errRest.Code, errRest)
		return
	}

	c.JSON(http.StatusOK, bidOutputList)
}
