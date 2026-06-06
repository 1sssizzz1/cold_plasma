package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type UploadService struct {
	dir string
}

type UploadResult struct {
	URL  string `json:"url"`
	MIME string `json:"mime"`
	Size int64  `json:"size"`
}

func NewUploadService(dir string) *UploadService {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		dir = "./uploads"
	}
	return &UploadService{dir: dir}
}

func (s *UploadService) ProcedureMedia(ctx context.Context, kind string, file multipart.File, header *multipart.FileHeader) (UploadResult, error) {
	kind = strings.TrimSpace(strings.ToLower(kind))
	if kind != "image" {
		return UploadResult{}, se(ErrValidation, "Поддерживается только загрузка картинок")
	}
	if header == nil || file == nil {
		return UploadResult{}, se(ErrValidation, "Файл обязателен")
	}
	maxSize := int64(20 << 20)
	if header.Size > maxSize {
		return UploadResult{}, se(ErrValidation, "Файл слишком большой")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExt(ext, []string{".jpg", ".jpeg", ".png", ".webp", ".heic", ".heif"}) {
		return UploadResult{}, se(ErrValidation, "Поддерживаются изображения jpg, png, webp, heic, heif")
	}

	tmpDir := filepath.Join(s.dir, "tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return UploadResult{}, fmt.Errorf("create upload tmp dir: %w", err)
	}
	id, err := randomName()
	if err != nil {
		return UploadResult{}, err
	}
	tmpPath := filepath.Join(tmpDir, id+ext)
	tmp, err := os.Create(tmpPath)
	if err != nil {
		return UploadResult{}, fmt.Errorf("create upload tmp file: %w", err)
	}
	if _, err := io.Copy(tmp, file); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return UploadResult{}, fmt.Errorf("save upload tmp file: %w", err)
	}
	_ = tmp.Close()
	defer os.Remove(tmpPath)

	return s.convertImage(ctx, tmpPath, id)
}

func (s *UploadService) convertImage(ctx context.Context, inputPath, id string) (UploadResult, error) {
	outDir := filepath.Join(s.dir, "procedures", "images")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return UploadResult{}, fmt.Errorf("create image dir: %w", err)
	}
	outPath := filepath.Join(outDir, id+".webp")
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "vips", "thumbnail", inputPath, outPath+"[Q=82]", "1600", "--size", "down")
	if out, err := cmd.CombinedOutput(); err != nil {
		return UploadResult{}, fmt.Errorf("convert image: %w: %s", err, strings.TrimSpace(string(out)))
	}
	info, err := os.Stat(outPath)
	if err != nil {
		return UploadResult{}, fmt.Errorf("stat image: %w", err)
	}
	return UploadResult{URL: "/uploads/procedures/images/" + id + ".webp", MIME: "image/webp", Size: info.Size()}, nil
}

func allowedExt(ext string, allowed []string) bool {
	for _, item := range allowed {
		if ext == item {
			return true
		}
	}
	return false
}

func randomName() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate upload name: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}
