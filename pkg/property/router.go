package property

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
	router.Use(net.VerifyJWT(this.cfg.JWTSecretKey))

	router.POST("", this.http.createProperty)
	router.GET("", this.http.getProperties)
	router.GET("/:id", this.http.getPropertyInfo)
	router.PATCH("/:id", this.http.updatePropertyInfo)
	router.POST("/:id/archive", this.http.archiveProperty)
	router.PUT("/:id/manager", this.http.assignPropertyManager)
	router.DELETE("/:id/manager", this.http.unassignPropertyManager)

	router.GET("/:id/units", this.http.getUnits)
	router.POST("/:id/units", this.http.createUnit)
	router.PATCH("/:id/units/:unitId", this.http.updateUnit)
	router.DELETE("/:id/units/:unitId", this.http.deleteUnit)

	router.POST("/managers", this.http.createPropertyManager)
	router.GET("/managers", this.http.getPropertyManagers)
	router.GET("/managers/:id", this.http.getPropertyManager)
	router.PATCH("/managers/:id", this.http.updatePropertyManager)
	router.DELETE("/managers/:id", this.http.removePropertyManager)
}
