package tenant

import "github.com/gin-gonic/gin"

type router struct {
	http *httpHandler
	path string
}

func CreateRouter(http *httpHandler, path string) *router {
	return &router{
		http: http,
		path: path,
	}
}

func (this *router) GetPath() string {
	return this.path
}

func (this *router) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/", this.http.createTenant)
	router.GET("/", this.http.getTenants)
	router.GET("/:id", this.http.getTenantInfo)
	router.PATCH("/:id", this.http.updateTenantInfo)
}
