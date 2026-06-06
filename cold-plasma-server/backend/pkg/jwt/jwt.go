package jwt

import (
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID   int    `json:"user_id"`
    Email    string `json:"email"`
    Name     string `json:"name"`
    IsAdmin  bool   `json:"is_admin"`
    jwt.RegisteredClaims
}

func GenerateToken(userID int, email, name string, isAdmin bool, secret string, expirySeconds int) (string, error) {
    claims := &Claims{
        UserID:  userID,
        Email:   email,
        Name:    name,
        IsAdmin: isAdmin,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expirySeconds) * time.Second)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

func ValidateToken(tokenStr, secret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Method)
        }
        return []byte(secret), nil
    })
    if err != nil {
        return nil, fmt.Errorf("token validation: %w", err)
    }

    claims, ok := token.Claims.(*Claims)
    if !ok {
        return nil, fmt.Errorf("cannot extract claims")
    }
    return claims, nil
}

