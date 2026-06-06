package handler

import (
	"net/http"
	"strconv"

	"cold-plasma-server/internal/service"
	"cold-plasma-server/internal/transport/middleware"

	"github.com/gin-gonic/gin"
)

type BonusHandler struct {
	bonus *service.BonusService
}

func NewBonusHandler(bonus *service.BonusService) *BonusHandler {
	return &BonusHandler{bonus: bonus}
}

func (h *BonusHandler) Balance(c *gin.Context) {
	uCtx, okUser := middleware.GetUser(c)
	if !okUser {
		fail(c, http.StatusUnauthorized, "Нужна авторизация")
		return
	}
	balance, err := h.bonus.Balance(c.Request.Context(), uCtx.UserID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"bonus_points": balance})
}

func (h *BonusHandler) Logs(c *gin.Context) {
	uCtx, okUser := middleware.GetUser(c)
	if !okUser {
		fail(c, http.StatusUnauthorized, "Нужна авторизация")
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	items, err := h.bonus.Logs(c.Request.Context(), uCtx.UserID, limit)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, items)
}

type spendReq struct {
	Amount  int    `json:"amount"`
	Comment string `json:"comment"`
}

func (h *BonusHandler) Spend(c *gin.Context) {
	uCtx, okUser := middleware.GetUser(c)
	if !okUser {
		fail(c, http.StatusUnauthorized, "Нужна авторизация")
		return
	}

	var req spendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	newBalance, err := h.bonus.Spend(c.Request.Context(), uCtx.UserID, req.Amount, req.Comment)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"bonus_points": newBalance})
}

