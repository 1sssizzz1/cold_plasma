package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type EmailVerifyClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Type   string `json:"type"` // email_verify
	jwt.RegisteredClaims
}

func GenerateEmailVerifyToken(userID int64, email, secret string, expirySeconds int) (string, error) {
	claims := &EmailVerifyClaims{
		UserID: userID,
		Email:  email,
		Type:   "email_verify",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expirySeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateEmailVerifyToken(tokenStr, secret string) (*EmailVerifyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &EmailVerifyClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Method)
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("token validation: %w", err)
	}
	if token == nil || !token.Valid {
		return nil, fmt.Errorf("token invalid")
	}

	claims, ok := token.Claims.(*EmailVerifyClaims)
	if !ok {
		return nil, fmt.Errorf("cannot extract claims")
	}
	if claims.Type != "email_verify" {
		return nil, fmt.Errorf("wrong token type")
	}
	return claims, nil
}

