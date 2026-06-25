package auction

import (
	"context"
	"fullcycle-auction_go/internal/infra/api/web/httperr"
	"fullcycle-auction_go/internal/infra/api/web/validation"
	"fullcycle-auction_go/internal/usecase/auction"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	auctionUseCase auction.UseCase
}

func New(auctionUseCase auction.UseCase) *Controller {
	return &Controller{
		auctionUseCase: auctionUseCase,
	}
}

func (ctrl *Controller) CreateAuction(c *gin.Context) {
	var auctionInputDTO auction.InputDTO

	if err := c.ShouldBindJSON(&auctionInputDTO); err != nil {
		restErr := validation.ValidateErr(err)

		c.JSON(restErr.Code, restErr)
		return
	}

	err := ctrl.auctionUseCase.CreateAuction(context.Background(), auctionInputDTO)
	if err != nil {
		restErr := httperr.ConvertError(err)

		c.JSON(restErr.Code, restErr)
		return
	}

	c.Status(http.StatusCreated)
}
