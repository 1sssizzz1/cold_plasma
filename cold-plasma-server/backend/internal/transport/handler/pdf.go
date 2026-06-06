package handler

import (
	"net/http"
	"strings"

	"cold-plasma-server/internal/service"
	"cold-plasma-server/internal/transport/middleware"

	"github.com/gin-gonic/gin"
)

type PDFHandler struct {
	pdf *service.PDFService
}

func NewPDFHandler(pdf *service.PDFService) *PDFHandler {
	return &PDFHandler{pdf: pdf}
}

func (h *PDFHandler) CareMemo(c *gin.Context) {
	name := strings.TrimSpace(c.Query("name"))
	if name == "" {
		if u, ok := middleware.GetUser(c); ok {
			name = u.Name
		}
	}

	b, err := h.pdf.CareMemo(c.Request.Context(), name)
	if err != nil {
		fail(c, http.StatusInternalServerError, "Не удалось сформировать PDF")
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline; filename=\"care-memo.pdf\"")
	c.Writer.WriteHeader(http.StatusOK)
	_, _ = c.Writer.Write(b)
}

