package common

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/patrickishaf/property-mgt/config"
)

var testKey = []byte("test-secret")

func signTestToken(t *testing.T, tokenType string, expiresIn time.Duration) string {
	t.Helper()
	claims := Claims{
		Actor:     Actor{ID: 1, Email: "a@b.c", Role: ActorRoleTenant},
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(testKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return token
}

func TestParseJWTExpired(t *testing.T) {
	token := signTestToken(t, TokenTypeAccess, -time.Minute)

	if _, err := ParseJWT(token, testKey, false); err == nil {
		t.Fatal("expected expired token to be rejected when allowExpired is false")
	}
	claims, err := ParseJWT(token, testKey, true)
	if err != nil {
		t.Fatalf("expected expired token to be accepted when allowExpired is true: %v", err)
	}
	if claims.Actor.ID != 1 {
		t.Fatalf("expected actor id 1, got %d", claims.Actor.ID)
	}
}

func TestParseJWTBadSignature(t *testing.T) {
	token := signTestToken(t, TokenTypeAccess, -time.Minute)
	if _, err := ParseJWT(token, []byte("wrong-secret"), true); err == nil {
		t.Fatal("expected token with bad signature to be rejected")
	}
}

func TestExtractActorFromJWTRejectsRefreshToken(t *testing.T) {
	if _, err := ExtractActorFromJWT(signTestToken(t, TokenTypeRefresh, time.Hour), testKey); err == nil {
		t.Fatal("expected refresh token to be rejected as an access token")
	}
	if _, err := ExtractActorFromJWT(signTestToken(t, TokenTypeAccess, time.Hour), testKey); err != nil {
		t.Fatalf("expected access token to be accepted: %v", err)
	}
}

func TestInviteJWTRoundTrip(t *testing.T) {
	cfg := config.AppConfig{JWTSecretKey: string(testKey), PwdResetExpirationSeconds: 60}
	token, err := GenerateInviteJWT(cfg, Actor{ID: 9, Role: ActorRoleTenant})
	if err != nil {
		t.Fatalf("failed to generate invite token: %v", err)
	}
	claims, err := ParseJWT(token, testKey, false)
	if err != nil {
		t.Fatalf("failed to parse invite token: %v", err)
	}
	if claims.TokenType != TokenTypeInvite || claims.Actor.ID != 9 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if _, err := ExtractActorFromJWT(token, testKey); err == nil {
		t.Fatal("expected invite token to be rejected as an access token")
	}
}
