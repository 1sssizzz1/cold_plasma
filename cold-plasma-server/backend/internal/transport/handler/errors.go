package handler

import (
	"errors"
	"net/http"

	"cold-plasma-server/internal/service"

	"github.com/gin-gonic/gin"
)

func handleServiceErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		fail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrUnauthorized):
		fail(c, http.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrForbidden):
		fail(c, http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrNotFound):
		fail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrConflict):
		fail(c, http.StatusConflict, err.Error())
	default:
		fail(c, http.StatusInternalServerError, "Внутренняя ошибка")
	}
}
