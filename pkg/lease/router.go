package lease

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
	// Runs the daily expiry job on demand so it can be exercised locally; it is
	// not authenticated and only exists in development.
	if this.cfg.Env == "development" {
		router.POST("/expiry/run", this.http.runExpiry)
	}

	authed := router.Group("", net.VerifyJWT(this.cfg.JWTSecretKey))
	authed.GET("", this.http.getLeases)
	authed.POST("", this.http.createLease)
	authed.GET("/installments", this.http.getInstallments)
	authed.PATCH("/installments/:id", this.http.updateInstallment)
	authed.GET("/:id", this.http.getLeaseById)
	authed.PATCH("/:id", this.http.updateLease)
	authed.GET("/:id/installments", this.http.getLeaseInstallments)
	authed.POST("/:id/installments", this.http.createLeaseInstallment)
}
