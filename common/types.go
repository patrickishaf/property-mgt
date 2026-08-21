package common

import "github.com/gin-gonic/gin"

type AppComponent interface {
	Run()
}

type AppRouter interface {
	GetPath() string
	RegisterRoutes(router *gin.RouterGroup)
}

type ServiceResponse struct {
	Data       any
	Err        error
	StatusCode int
}
