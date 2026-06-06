package handler

import (
	"net/http"

	"cold-plasma-server/internal/service"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	uploads *service.UploadService
}

func NewUploadHandler(uploads *service.UploadService) *UploadHandler {
	return &UploadHandler{uploads: uploads}
}

func (h *UploadHandler) ProcedureMedia(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(320 << 20); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный multipart-запрос")
		return
	}
	kind := c.PostForm("kind")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		fail(c, http.StatusBadRequest, "Файл обязателен")
		return
	}
	defer file.Close()
	out, err := h.uploads.ProcedureMedia(c.Request.Context(), kind, file, header)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, out)
}
