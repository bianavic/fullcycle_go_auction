package web

import (
	"net/http"

	auctioncontroller "fullcycle-auction_go/internal/infra/api/web/controller/auction"
	"fullcycle-auction_go/internal/infra/api/web/controller/bid"
	"fullcycle-auction_go/internal/infra/api/web/controller/user"

	"github.com/gin-gonic/gin"
)

// NewRouter monta as rotas HTTP da aplicação a partir dos controllers já
// inicializados. Extraído de cmd/auction/main.go para poder ser exercitado
// por testes E2E sem duplicar o registro de rotas.
func NewRouter(
	userController *user.Controller,
	bidController *bid.Controller,
	auctionController *auctioncontroller.Controller,
) *gin.Engine {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/auction", auctionController.FindAuctions)
	router.GET("/auction/:auctionId", auctionController.FindAuctionByID)
	router.POST("/auction", auctionController.CreateAuction)
	router.GET("/auction/winner/:auctionId", auctionController.FindWinningBidByAuctionID)
	router.POST("/bid", bidController.CreateBid)
	router.GET("/bid/:auctionId", bidController.FindBidByAuctionID)
	router.GET("/bid/winner/:auctionId", bidController.FindWinningBidByAuctionID)
	router.GET("/user/:userId", userController.FindUserByID)

	return router
}
