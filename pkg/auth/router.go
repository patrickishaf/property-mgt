package auth

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
	router.POST("/accounts", this.http.createOwnerAccount)
	router.GET("/user", this.http.getAuthenticatedUser)
	router.POST("/login", this.http.loginWithPassword)
	router.POST("/logout", this.http.logout)
	router.POST("/otp", this.http.sendOTP)
	router.POST("/otp-verify", this.http.verifyOTP)
}
