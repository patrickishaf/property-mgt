package settings

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
	router.GET("/:userId", this.http.getUserSettings)
	router.PATCH("/profile/:userId", this.http.updateProfileSettings)
	router.PATCH("/preferences/:userId", this.http.updatePreferences)
}
