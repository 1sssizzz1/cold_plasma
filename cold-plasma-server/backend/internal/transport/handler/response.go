package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type apiError struct {
	Message string `json:"message"`
}

func ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

func fail(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": apiError{Message: message}})
}

