package bid

import (
	"fullcycle-auction_go/internal/infra/api/web/httperr"
	"fullcycle-auction_go/internal/infra/api/web/validation"
	biduc "fullcycle-auction_go/internal/usecase/bid"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	bid biduc.UseCase
}

func New(bidUseCase biduc.UseCase) *Controller {
	return &Controller{
		bid: bidUseCase,
	}
}

func (u *Controller) CreateBid(c *gin.Context) {
	var bidInputDTO biduc.InputDTO

	if err := c.ShouldBindJSON(&bidInputDTO); err != nil {
		restErr := validation.ValidateErr(err)

		c.JSON(restErr.Code, restErr)
		return
	}

	err := u.bid.CreateBid(c.Request.Context(), bidInputDTO)
	if err != nil {
		restErr := httperr.ConvertError(err)

		c.JSON(restErr.Code, restErr)
		return
	}

	c.Status(http.StatusCreated)
}
