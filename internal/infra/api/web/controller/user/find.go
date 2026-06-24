package user

import (
	"context"
	"fullcycle-auction_go/internal/infra/api/web/httperr"
	"fullcycle-auction_go/internal/infra/api/web/validation"
	"fullcycle-auction_go/internal/usecase/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	user user.UseCase
}

func New(userUseCase user.UseCase) *Controller {
	return &Controller{
		user: userUseCase,
	}
}

func (u *Controller) FindUserByID(c *gin.Context) {
	userID := c.Param("userId")

	if errRest := validation.ValidateUUID(userID, "userId"); errRest != nil {
		c.JSON(errRest.Code, errRest)
		return
	}

	userData, err := u.user.FindUserByID(context.Background(), userID)
	if err != nil {
		errRest := httperr.ConvertError(err)
		c.JSON(errRest.Code, errRest)
		return
	}

	c.JSON(http.StatusOK, userData)
}
