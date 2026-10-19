package messaging

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
	router.GET("/recipients", this.http.getMessageRecipients)
	router.GET("/messages/:id", this.http.getMessageInfo)
	router.GET("/messages", this.http.getMessages)
	router.GET("/templates", this.http.getTemplates)
	router.POST("/messages", this.http.sendMessage)
}
