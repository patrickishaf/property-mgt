package common

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/patrickishaf/property-mgt/config"
)

type Claims struct {
	Actor Actor `json:"actor"`
	jwt.RegisteredClaims
}

func ExtractActorFromJWT(token string, jwtKey []byte) (*Actor, error) {
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !tkn.Valid {
		return nil, errors.New(ErrInvalidRequest)
	}
	return &claims.Actor, nil
}

func GenerateJWT(cfg config.AppConfig, actor Actor) (string, error) {
	expirationTime := time.Now().Add(time.Duration(cfg.AccessTokenExpirationMinutes) * time.Minute)
	claims := Claims{
		Actor: actor,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecretKey))
}
