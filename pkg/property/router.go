package property

import "github.com/gin-gonic/gin"

type router struct {
	http *httpHandler
	path string
}

func CreateRouter(httpHandler *httpHandler, path string) *router {
	return &router{
		http: httpHandler,
		path: path,
	}
}

func (this *router) GetPath() string {
	return this.path
}

func (this *router) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/managers", this.http.createPropertyManager)
	router.GET("/", this.http.getProperties)
	router.GET("/:id", this.http.getPropertyInfo)
	router.GET("/managers/:id", this.http.getPropertyManager)
	router.GET("/managers", this.http.getPropertyManagers)
	router.DELETE("/managers/:id", this.http.removePropertyManager)
	router.PATCH("/:id", this.http.updatePropertyInfo)
	router.PATCH("/managers/:id", this.http.updatePropertyManager)
}
