package common

import "github.com/gin-gonic/gin"

const ActorRoleTenant = "tenant"

type Actor struct {
	ID          uint   `json:"id"`
	Email       string `json:"email"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	PhoneNumber string `json:"phonenumber"`
	Role        string `json:"role"`
}

type AdminRole string

const (
	AdminRoleOwner   AdminRole = "owner"
	AdminRoleManager AdminRole = "property_manager"
)

func (ar AdminRole) ToString() string {
	return string(ar)
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
	Err        string
	StatusCode int
}
