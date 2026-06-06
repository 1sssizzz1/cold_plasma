package middleware

import (
	"net/http"
	"strings"

	appjwt "cold-plasma-server/pkg/jwt"

	"github.com/gin-gonic/gin"
)

const userCtxKey = "user_ctx"

type UserCtx struct {
	UserID  int64  `json:"user_id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"is_admin"`
}

func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearer(c.GetHeader("Authorization"))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"message": "Нужна авторизация"}})
			return
		}

		claims, err := appjwt.ValidateToken(token, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"message": "Токен недействителен"}})
			return
		}

		c.Set(userCtxKey, UserCtx{
			UserID:  int64(claims.UserID),
			Email:   claims.Email,
			Name:    claims.Name,
			IsAdmin: claims.IsAdmin,
		})
		c.Next()
	}
}

func AuthOptional(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearer(c.GetHeader("Authorization"))
		if token == "" {
			c.Next()
			return
		}
		claims, err := appjwt.ValidateToken(token, secret)
		if err == nil && claims != nil {
			c.Set(userCtxKey, UserCtx{
				UserID:  int64(claims.UserID),
				Email:   claims.Email,
				Name:    claims.Name,
				IsAdmin: claims.IsAdmin,
			})
		}
		c.Next()
	}
}

func GetUser(c *gin.Context) (UserCtx, bool) {
	v, ok := c.Get(userCtxKey)
	if !ok {
		return UserCtx{}, false
	}
	u, ok := v.(UserCtx)
	return u, ok
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := GetUser(c)
		if !ok || !u.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": gin.H{"message": "Нужны права администратора"}})
			return
		}
		c.Next()
	}
}

func extractBearer(h string) string {
	h = strings.TrimSpace(h)
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if strings.ToLower(strings.TrimSpace(parts[0])) != "bearer" {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
