package common

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/patrickishaf/property-mgt/config"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeInvite  = "invite"
	TokenTypeRefresh = "refresh"
)

type Claims struct {
	Actor     Actor  `json:"actor"`
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

// ParseJWT verifies the token's signature and returns its claims. When allowExpired
// is true, a token whose only validation failure is expiry is still accepted.
func ParseJWT(token string, jwtKey []byte, allowExpired bool) (*Claims, error) {
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		var ve *jwt.ValidationError
		if allowExpired && errors.As(err, &ve) && ve.Errors == jwt.ValidationErrorExpired {
			return claims, nil
		}
		return nil, err
	}
	if !tkn.Valid {
		return nil, errors.New(ErrInvalidRequest)
	}
	return claims, nil
}

func ExtractActorFromJWT(token string, jwtKey []byte) (*Actor, error) {
	claims, err := ParseJWT(token, jwtKey, false)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeAccess {
		return nil, errors.New(ErrInvalidToken)
	}
	return &claims.Actor, nil
}

func GenerateJWT(cfg config.AppConfig, actor Actor) (string, error) {
	expirationTime := time.Now().Add(time.Duration(cfg.AccessTokenExpirationMinutes) * time.Minute)
	claims := Claims{
		Actor:     actor,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecretKey))
}

// GenerateInviteJWT creates a short-lived token that lets a new tenant set their password.
func GenerateInviteJWT(cfg config.AppConfig, actor Actor) (string, error) {
	expirationTime := time.Now().Add(time.Duration(cfg.PwdResetExpirationSeconds) * time.Second)
	claims := Claims{
		Actor:     actor,
		TokenType: TokenTypeInvite,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecretKey))
}
