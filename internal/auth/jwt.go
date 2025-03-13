package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthenticator struct {
	keys EddsaKeys
	aud  string
	iss  string
}

func NewJWTAutheticatorWithEddsa(keys EddsaKeys, aud, iss string) *JWTAuthenticator {
	return &JWTAuthenticator{
		keys: keys,
		aud:  aud,
		iss:  iss,
	}
}

func (a *JWTAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

	tokenString, err := token.SignedString(a.keys.Private)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *JWTAuthenticator) ValidateTokenAndParse(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return a.keys.Public, nil
	}, jwt.WithExpirationRequired(), jwt.WithAudience(a.aud), jwt.WithIssuer(a.iss))
	if err != nil {
		return nil, err
	} else if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
		return nil, errors.New("unexpected jwt signing method detected")
	}
	return token, nil
}
