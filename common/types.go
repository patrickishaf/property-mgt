package common

import "github.com/gin-gonic/gin"

type Actor struct {
	ID          uint     `json:"id"`
	Email       string   `json:"email"`
	FirstName   string   `json:"firstName"`
	LastName    string   `json:"lastName"`
	PhoneNumber string   `json:"phonenumber"`
	Roles       []string `json:"roles"`
}

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
