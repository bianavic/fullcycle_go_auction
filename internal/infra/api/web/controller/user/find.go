package user

import (
	"context"
	"fullcycle-auction_go/configuration/httperr"
	"fullcycle-auction_go/internal/usecase/user"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	if err := uuid.Validate(userID); err != nil {
		errRest := httperr.NewBadRequestError("Invalid fields", httperr.Causes{
			Field:   "userId",
			Message: "Invalid UUID value",
		})

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
