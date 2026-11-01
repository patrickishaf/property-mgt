package settings

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
	router.Use(net.VerifyJWT(this.cfg.JWTSecretKey))
	router.GET("/:userId", this.http.getUserSettings)
	router.PATCH("/profile/:userId", this.http.updateProfileSettings)
	router.PATCH("/preferences/:userId", this.http.updatePreferences)
}
