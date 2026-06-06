package handler

import (
	"net/http"
	"strconv"

	"cold-plasma-server/internal/repository"
	"cold-plasma-server/internal/service"
	"cold-plasma-server/internal/transport/middleware"

	"github.com/gin-gonic/gin"
)

type ProcedureHandler struct {
	procedures *service.ProcedureService
}

func NewProcedureHandler(procedures *service.ProcedureService) *ProcedureHandler {
	return &ProcedureHandler{procedures: procedures}
}

func (h *ProcedureHandler) List(c *gin.Context) {
	items, err := h.procedures.List(c.Request.Context())
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, items)
}

func (h *ProcedureHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(c, http.StatusBadRequest, "Некорректный id")
		return
	}
	p, err := h.procedures.Get(c.Request.Context(), id)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, p)
}

func (h *ProcedureHandler) AdminList(c *gin.Context) {
	items, err := h.procedures.ListAll(c.Request.Context())
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, items)
}

type procedureReq struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	DurationMins int    `json:"duration_mins"`
	Price        int    `json:"price"`
	BonusEarned  int    `json:"bonus_earned"`
	Category     string `json:"category"`
	ImageURL     string `json:"image_url"`
	VideoURL     string `json:"video_url"`
	Services     string `json:"services"`
	DurationStr  string `json:"duration_str"`
	Popular      bool   `json:"popular"`
	IsActive     bool   `json:"is_active"`
}

func (h *ProcedureHandler) AdminCreate(c *gin.Context) {
	var req procedureReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	item, err := h.procedures.Save(c.Request.Context(), 0, procedureParams(req))
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	created(c, item)
}

func (h *ProcedureHandler) AdminUpdate(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req procedureReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	item, err := h.procedures.Save(c.Request.Context(), id, procedureParams(req))
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, item)
}

func (h *ProcedureHandler) AdminDelete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.procedures.Delete(c.Request.Context(), id); err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"ok": true})
}

func (h *ProcedureHandler) Reviews(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	items, err := h.procedures.ListReviews(c.Request.Context(), id)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, items)
}

func (h *ProcedureHandler) AdminReviews(c *gin.Context) {
	items, err := h.procedures.ListAllReviews(c.Request.Context())
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, items)
}

type reviewReq struct {
	ProcedureID int64  `json:"procedure_id"`
	Rating      int    `json:"rating"`
	Text        string `json:"text"`
}

func (h *ProcedureHandler) CreateReview(c *gin.Context) {
	uCtx, okUser := middleware.GetUser(c)
	if !okUser {
		fail(c, http.StatusUnauthorized, "Нужна авторизация")
		return
	}
	var req reviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	item, err := h.procedures.AddReview(c.Request.Context(), uCtx.UserID, req.ProcedureID, req.Rating, req.Text)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, item)
}

func (h *ProcedureHandler) AdminDeleteReview(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.procedures.DeleteReview(c.Request.Context(), id); err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"ok": true})
}

func procedureParams(req procedureReq) repository.ProcedureParams {
	return repository.ProcedureParams{
		Title:        req.Title,
		Description:  req.Description,
		DurationMins: req.DurationMins,
		Price:        req.Price,
		BonusEarned:  req.BonusEarned,
		Category:     req.Category,
		ImageURL:     req.ImageURL,
		VideoURL:     req.VideoURL,
		Services:     req.Services,
		DurationStr:  req.DurationStr,
		Popular:      req.Popular,
		IsActive:     req.IsActive,
	}
}
