package handler

import (
	"net/url"
	"strconv"
	"time"

	"cold-plasma-server/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	admin *service.AdminService
	bonus *service.BonusService
}

func NewAdminHandler(admin *service.AdminService, bonus *service.BonusService) *AdminHandler {
	return &AdminHandler{admin: admin, bonus: bonus}
}

func (h *AdminHandler) ChatLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	items, err := h.admin.ChatLogs(c.Request.Context(), limit)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, items)
}

func (h *AdminHandler) ChatSessions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	items, err := h.admin.ChatSessions(c.Request.Context(), limit)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, items)
}

func (h *AdminHandler) DeleteChatSession(c *gin.Context) {
	sessionID, err := url.PathUnescape(c.Param("id"))
	if err != nil {
		fail(c, 400, "Некорректный id сессии")
		return
	}
	if err := h.admin.DeleteChatSession(c.Request.Context(), sessionID); err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"ok": true})
}

func (h *AdminHandler) BookingRequests(c *gin.Context) {
	items, err := h.admin.BookingRequests(c.Request.Context())
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, items)
}

func (h *AdminHandler) ActiveBookings(c *gin.Context) {
	items, err := h.admin.ActiveBookings(c.Request.Context())
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, items)
}

func (h *AdminHandler) CompletedBookings(c *gin.Context) {
	items, err := h.admin.CompletedBookings(c.Request.Context())
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, items)
}

type confirmBookingReq struct {
	DateTime string `json:"datetime"`
}

func (h *AdminHandler) ConfirmBooking(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req confirmBookingReq
	_ = c.ShouldBindJSON(&req)
	var dt time.Time
	if req.DateTime != "" {
		parsed, err := time.Parse(time.RFC3339, req.DateTime)
		if err != nil {
			fail(c, 400, "Некорректный datetime")
			return
		}
		dt = parsed
	}
	if err := h.admin.ConfirmBooking(c.Request.Context(), id, dt); err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"ok": true})
}

func (h *AdminHandler) CompleteBooking(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.admin.CompleteBooking(c.Request.Context(), id); err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"ok": true})
}

func (h *AdminHandler) RescheduleBooking(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req confirmBookingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "Некорректный JSON")
		return
	}
	parsed, err := time.Parse(time.RFC3339, req.DateTime)
	if err != nil {
		fail(c, 400, "Некорректный datetime")
		return
	}
	if err := h.admin.RescheduleBooking(c.Request.Context(), id, parsed); err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"ok": true})
}

func (h *AdminHandler) DeleteBooking(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.admin.DeleteBooking(c.Request.Context(), id); err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"ok": true})
}

func (h *AdminHandler) Calendar(c *gin.Context) {
	from, err := time.Parse(time.RFC3339, c.Query("from"))
	if err != nil {
		fail(c, 400, "Некорректная дата from (RFC3339)")
		return
	}
	to, err := time.Parse(time.RFC3339, c.Query("to"))
	if err != nil {
		fail(c, 400, "Некорректная дата to (RFC3339)")
		return
	}
	data, err := h.admin.Calendar(c.Request.Context(), from, to)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, data)
}

type createNoteReq struct {
	StartAt string `json:"start_at"`
	EndAt   string `json:"end_at"`
	Title   string `json:"title"`
}

func (h *AdminHandler) CreateNote(c *gin.Context) {
	var req createNoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "Некорректный JSON")
		return
	}
	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		fail(c, 400, "Некорректная дата начала (RFC3339)")
		return
	}
	endAt, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		fail(c, 400, "Некорректная дата конца (RFC3339)")
		return
	}
	note, err := h.admin.CreateNote(c.Request.Context(), startAt, endAt, req.Title)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	created(c, gin.H{"note": note})
}

func (h *AdminHandler) DeleteNote(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.admin.DeleteNote(c.Request.Context(), id); err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"ok": true})
}

func (h *AdminHandler) NotificationSettings(c *gin.Context) {
	settings, err := h.admin.NotificationSettings(c.Request.Context())
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, settings)
}

func (h *AdminHandler) MasterProfile(c *gin.Context) {
	profile, err := h.admin.MasterProfile(c.Request.Context())
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, profile)
}

func (h *AdminHandler) UpdateMasterProfile(c *gin.Context) {
	var req service.MasterProfile
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "Некорректный JSON")
		return
	}
	profile, err := h.admin.UpdateMasterProfile(c.Request.Context(), req)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, profile)
}

func (h *AdminHandler) UpdateNotificationSettings(c *gin.Context) {
	var req service.AdminNotificationSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "Некорректный JSON")
		return
	}
	settings, err := h.admin.UpdateNotificationSettings(c.Request.Context(), req)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, settings)
}

type awardBonusReq struct {
	Phone   string `json:"phone"`
	Amount  int    `json:"amount"`
	Comment string `json:"comment"`
}

func (h *AdminHandler) BonusUserByPhone(c *gin.Context) {
	user, err := h.bonus.FindByPhone(c.Request.Context(), c.Query("phone"))
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"user": user, "bonus_points": user.BonusPoints})
}

func (h *AdminHandler) SpendBonus(c *gin.Context) {
	var req awardBonusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "Некорректный JSON")
		return
	}
	user, err := h.bonus.SpendByPhone(c.Request.Context(), req.Phone, req.Amount, req.Comment)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"user": user, "bonus_points": user.BonusPoints})
}

func (h *AdminHandler) AwardBonus(c *gin.Context) {
	var req awardBonusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "Некорректный JSON")
		return
	}
	user, err := h.bonus.AwardByPhone(c.Request.Context(), req.Phone, req.Amount, req.Comment)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"user": user, "bonus_points": user.BonusPoints})
}
