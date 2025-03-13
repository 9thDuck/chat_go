package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Authenticator interface {
	GenerateToken(claims jwt.Claims) (string, error)
	ValidateTokenAndParse(tokenString string) (*jwt.Token, error)
}

type ExpiryDurations struct {
	Access  time.Duration
	Refresh time.Duration
}
