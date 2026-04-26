package handler

import (
	"net/http"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/storage"
	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	storage *storage.S3Storage
}

func NewUploadHandler(s *storage.S3Storage) *UploadHandler {
	return &UploadHandler{storage: s}
}

// UploadProductImage загружает изображение товара в S3
// POST /api/upload/product-image
// multipart/form-data: file=<image>
func (h *UploadHandler) UploadProductImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Ограничение размера 5 MB
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large, max 5MB"})
		return
	}

	contentType := file.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}
	if !allowedTypes[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only jpeg, png, webp allowed"})
		return
	}

	url, err := h.storage.UploadFile(c.Request.Context(), file, "products")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}
