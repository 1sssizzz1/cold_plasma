package handler

import (
	"crypto/subtle"
	"net/http"

	tgservice "cold-plasma-server/internal/telegram/service"
	"cold-plasma-server/internal/transport/middleware"

	"github.com/gin-gonic/gin"
)

type TelegramHandler struct {
	telegram *tgservice.Service
	secret   string
}

func NewTelegramHandler(telegram *tgservice.Service, secret string) *TelegramHandler {
	return &TelegramHandler{telegram: telegram, secret: secret}
}

func (h *TelegramHandler) CreateLink(c *gin.Context) {
	uCtx, okUser := middleware.GetUser(c)
	if !okUser {
		fail(c, http.StatusUnauthorized, "Нужна авторизация")
		return
	}
	out, err := h.telegram.CreateLink(c.Request.Context(), uCtx.UserID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, out)
}

func (h *TelegramHandler) Status(c *gin.Context) {
	uCtx, okUser := middleware.GetUser(c)
	if !okUser {
		fail(c, http.StatusUnauthorized, "Нужна авторизация")
		return
	}
	out, err := h.telegram.LinkStatus(c.Request.Context(), uCtx.UserID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, out)
}

func (h *TelegramHandler) Webhook(c *gin.Context) {
	// Секрет обязателен: без него webhook остаётся публичной точкой,
	// через которую можно слать поддельные апдейты Telegram.
	if h.secret == "" {
		fail(c, http.StatusServiceUnavailable, "Telegram webhook не настроен")
		return
	}
	got := c.GetHeader("X-Telegram-Bot-Api-Secret-Token")
	if subtle.ConstantTimeCompare([]byte(got), []byte(h.secret)) != 1 {
		fail(c, http.StatusUnauthorized, "Некорректный Telegram secret")
		return
	}
	var update tgservice.Update
	if err := c.ShouldBindJSON(&update); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	if err := h.telegram.HandleUpdate(c.Request.Context(), update); err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"ok": true})
}
