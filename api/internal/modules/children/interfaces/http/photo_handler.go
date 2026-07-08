package httpchild

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/children/application"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/tenant"
)

const (
	maxUploadSize = 5 * 1024 * 1024 // 5 MB
)

var allowedMimeTypes = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
}

type PhotoUseCases struct {
	Upload *application.UploadPhoto
	Remove *application.RemovePhoto
}

func (h *Handler) uploadPhotoHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	file, err := c.FormFile("photo")
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Missing photo file.", []map[string]string{{"field": "photo", "message": "required"}})
		return
	}

	if file.Size > maxUploadSize {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "File exceeds maximum size of 5 MB.", []map[string]string{{"field": "photo", "message": "must be under 5 MB"}})
		return
	}

	f, err := file.Open()
	if err != nil {
		httpserver.WriteError(c, http.StatusInternalServerError, "internal_error", "Failed to read uploaded file.", nil)
		return
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		httpserver.WriteError(c, http.StatusInternalServerError, "internal_error", "Failed to read file content.", nil)
		return
	}
	buf = buf[:n]

	mime := mimetype.Detect(buf)
	ext, ok := allowedMimeTypes[mime.String()]
	if !ok {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid file type. Only JPEG and PNG are accepted.", []map[string]string{{"field": "photo", "message": "must be JPEG or PNG"}})
		return
	}

	if _, seekErr := f.Seek(0, io.SeekStart); seekErr != nil {
		httpserver.WriteError(c, http.StatusInternalServerError, "internal_error", "Failed to process file.", nil)
		return
	}

	result, err := h.uploadPhoto.Execute(c.Request.Context(), actor, c.Param("child_id"), f, ext)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"photo_url": result.PhotoURL})
}

func (h *Handler) removePhotoHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	if err := h.removePhoto.Execute(c.Request.Context(), actor, c.Param("child_id")); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"photo_url": nil})
}

func (h *Handler) getPhotoHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	child, err := h.getChild.Execute(c.Request.Context(), actor, c.Param("child_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}

	if child.ProfilePhotoPath == nil {
		httpserver.WriteError(c, http.StatusNotFound, "not_found", "No photo found for this child.", nil)
		return
	}

	ext := strings.ToLower(filepath.Ext(*child.ProfilePhotoPath))
	contentType := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	}

	c.Header("Cache-Control", "private, max-age=3600")
	c.File(*child.ProfilePhotoPath)
	c.Header("Content-Type", contentType)
}

func (h *Handler) getPhotoURLEndpoint(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	child, err := h.getChild.Execute(c.Request.Context(), actor, c.Param("child_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}

	if child.ProfilePhotoPath == nil {
		c.JSON(http.StatusOK, gin.H{"photo_url": nil})
		return
	}

	photoURL := fmt.Sprintf("/api/v1/children/%s/photo", c.Param("child_id"))
	c.JSON(http.StatusOK, gin.H{"photo_url": photoURL})
}
