package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type UploadService struct {
	uploadDir string
	baseURL   string
}

var allowedMimeTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"application/pdf": true,
}

const maxFileSize = 5 * 1024 * 1024 // 5MB

func NewUploadService(uploadDir, baseURL string) *UploadService {
	return &UploadService{uploadDir: uploadDir, baseURL: baseURL}
}

func (s *UploadService) UploadFile(file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	if header.Size > maxFileSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of 5MB")
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedMimeTypes[contentType] {
		return "", fmt.Errorf("file type %s is not allowed, accepted: jpeg, png, webp, pdf", contentType)
	}

	ext := strings.ToLower(path.Ext(header.Filename))
	if ext == "" {
		ext = ".bin"
	}

	fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	destDir := filepath.Join(s.uploadDir, folder)

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	destPath := filepath.Join(destDir, fileName)
	dst, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	url := fmt.Sprintf("%s/uploads/%s/%s", s.baseURL, folder, fileName)
	return url, nil
}
