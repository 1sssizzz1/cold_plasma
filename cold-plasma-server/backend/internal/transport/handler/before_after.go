package handler

import (
	"net/http"
	"strconv"

	"cold-plasma-server/internal/repository"
	"cold-plasma-server/internal/service"

	"github.com/gin-gonic/gin"
)

type BeforeAfterHandler struct {
	results *service.BeforeAfterService
}

func NewBeforeAfterHandler(results *service.BeforeAfterService) *BeforeAfterHandler {
	return &BeforeAfterHandler{results: results}
}

type beforeAfterReq struct {
	ProcedureID *int64 `json:"procedure_id"`
	Procedure   string `json:"procedure"`
	Title       string `json:"title"`
	Description string `json:"description"`
	BeforeURL   string `json:"before_url"`
	AfterURL    string `json:"after_url"`
	IsFeatured  bool   `json:"is_featured"`
	SortOrder   int    `json:"sort_order"`
	IsActive    bool   `json:"is_active"`
}

func (h *BeforeAfterHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	items, err := h.results.List(c.Request.Context(), limit)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, items)
}

func (h *BeforeAfterHandler) AdminList(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	items, err := h.results.ListAll(c.Request.Context(), limit)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, items)
}

func (h *BeforeAfterHandler) AdminCreate(c *gin.Context) {
	var req beforeAfterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	item, err := h.results.Save(c.Request.Context(), 0, beforeAfterParams(req))
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	created(c, item)
}

func (h *BeforeAfterHandler) AdminUpdate(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req beforeAfterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	item, err := h.results.Save(c.Request.Context(), id, beforeAfterParams(req))
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, item)
}

func (h *BeforeAfterHandler) AdminDelete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.results.Delete(c.Request.Context(), id); err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"ok": true})
}

func beforeAfterParams(req beforeAfterReq) repository.BeforeAfterParams {
	return repository.BeforeAfterParams{
		ProcedureID: req.ProcedureID,
		Procedure:   req.Procedure,
		Title:       req.Title,
		Description: req.Description,
		BeforeURL:   req.BeforeURL,
		AfterURL:    req.AfterURL,
		IsFeatured:  req.IsFeatured,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
	}
}
