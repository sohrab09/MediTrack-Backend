package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenLifetime = 24 * time.Hour

func GenerateJWT(email, secret string) (string, error) {
	if email == "" {
		return "", fmt.Errorf("email is required for JWT")
	}

	if len(secret) < 32 {
		return "", fmt.Errorf("JWT secret must contain at least 32 bytes")
	}

	now := time.Now()

	claims := jwt.MapClaims{
		"email": email,
		"iat":   now.Unix(),
		"exp":   now.Add(tokenLifetime).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}

	return signedToken, nil
}
