package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/patrickishaf/property-mgt/common"
)

type createOwnerAccountReq struct {
	Email       string `json:"email"`
	FirstName   string `json:"fname"`
	LastName    string `json:"lname"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phoneNumber"`
}

type jwtPayload struct {
	Actor common.Actor
	jwt.RegisteredClaims
}

type passwordLoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Type     string `json:"accountType"`
}
