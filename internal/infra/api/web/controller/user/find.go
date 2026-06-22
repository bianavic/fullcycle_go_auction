package user

import (
	"context"
	"fullcycle-auction_go/configuration/rest_err"
	"fullcycle-auction_go/internal/usecase/user"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController struct {
	user user.UseCase
}

func New(userUseCase user.UseCase) *UserController {
	return &UserController{
		user: userUseCase,
	}
}

func (u *UserController) FindUserByID(c *gin.Context) {
	userID := c.Param("userId")

	if err := uuid.Validate(userID); err != nil {
		errRest := rest_err.NewBadRequestError("Invalid fields", rest_err.Causes{
			Field:   "userId",
			Message: "Invalid UUID value",
		})

		c.JSON(errRest.Code, errRest)
		return
	}

	userData, err := u.user.FindUserByID(context.Background(), userID)
	if err != nil {
		errRest := rest_err.ConvertError(err)
		c.JSON(errRest.Code, errRest)
		return
	}

	c.JSON(http.StatusOK, userData)
}
