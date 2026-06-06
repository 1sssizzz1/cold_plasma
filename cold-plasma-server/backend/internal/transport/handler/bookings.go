package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cold-plasma-server/internal/service"
	"cold-plasma-server/internal/transport/middleware"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	bookings *service.BookingService
}

func NewBookingHandler(bookings *service.BookingService) *BookingHandler {
	return &BookingHandler{bookings: bookings}
}

type createBookingReq struct {
	ProcedureID    int64    `json:"procedure_id"`
	DateTime       string   `json:"datetime"`  // legacy RFC3339
	DateTimes      []string `json:"datetimes"` // RFC3339 list
	Comment        string   `json:"comment"`
	BonusUsed      int      `json:"bonus_used"`
	NotifySMS      bool     `json:"notify_sms"`
	NotifyTelegram bool     `json:"notify_telegram"`
}

func (h *BookingHandler) Create(c *gin.Context) {
	uCtx, okUser := middleware.GetUser(c)
	if !okUser {
		fail(c, http.StatusUnauthorized, "Нужна авторизация")
		return
	}

	var req createBookingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	dates, err := parseRFC3339List(req.DateTimes, req.DateTime)
	if err != nil {
		fail(c, http.StatusBadRequest, "Некорректные даты (нужен RFC3339, например 2026-05-14T15:30:00+03:00)")
		return
	}

	booking, balance, err := h.bookings.Create(
		c.Request.Context(),
		uCtx.UserID,
		req.ProcedureID,
		dates,
		req.Comment,
		req.BonusUsed,
		req.NotifySMS,
		req.NotifyTelegram,
	)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	created(c, gin.H{"booking": booking, "bonus_balance": balance})
}

func parseRFC3339List(items []string, legacy string) ([]time.Time, error) {
	if len(items) == 0 && strings.TrimSpace(legacy) != "" {
		items = []string{legacy}
	}
	out := make([]time.Time, 0, len(items))
	for _, item := range items {
		t, err := parseRFC3339(item)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

func (h *BookingHandler) ListMy(c *gin.Context) {
	uCtx, okUser := middleware.GetUser(c)
	if !okUser {
		fail(c, http.StatusUnauthorized, "Нужна авторизация")
		return
	}

	items, err := h.bookings.ListMy(c.Request.Context(), uCtx.UserID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, items)
}

func parseRFC3339(v string) (time.Time, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}, errors.New("empty datetime")
	}
	// Строгий RFC3339
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t, nil
	}
	// Иногда фронт отправляет без секунд
	if t, err := time.Parse("2006-01-02T15:04-07:00", v); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid datetime")
}
