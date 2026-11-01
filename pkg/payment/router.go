package payment

import (
	"github.com/gin-gonic/gin"
	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/net"
)

type router struct {
	cfg  *config.AppConfig
	http *httpHandler
	path string
}

func CreateRouter(cfg *config.AppConfig, http *httpHandler, path string) *router {
	return &router{
		cfg:  cfg,
		http: http,
		path: path,
	}
}

func (this *router) GetPath() string {
	return this.path
}

func (this *router) RegisterRoutes(router *gin.RouterGroup) {
	// Stands in for the payment provider confirming a checkout until Bachs is
	// integrated, so it is not authenticated and only exists in development.
	if this.cfg.Env == "development" {
		router.POST("/checkout/:reference/complete", this.http.completeCheckout)
	}

	authed := router.Group("", net.VerifyJWT(this.cfg.JWTSecretKey))
	authed.POST("", this.http.createPayment)
	authed.GET("", this.http.getPayments)
	authed.GET("/:id", this.http.getPaymentInfo)
	authed.GET("/stats", this.http.getTotalPaymentStats)
	authed.POST("/checkout", this.http.initiateCheckout)
}
