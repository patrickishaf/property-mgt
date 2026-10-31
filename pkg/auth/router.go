package auth

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

func CreateRouter(cfg *config.AppConfig, httpHandler *httpHandler, path string) *router {
	return &router{
		cfg:  cfg,
		http: httpHandler,
		path: path,
	}
}

func (this *router) GetPath() string {
	return this.path
}

func (this *router) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/accounts", this.http.createOwnerAccount)
	router.GET("/user", net.VerifyJWT(this.cfg.JWTSecretKey), this.http.getAuthenticatedUser)
	router.POST("/login", this.http.loginWithPassword)
	router.POST("/token/refresh", this.http.refreshAuthToken)
}
