package handler

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/openlankapay/openlankapay/pkg/auth"
)

// AdminUploadHandler handles file uploads for admin assets (logos, etc).
type AdminUploadHandler struct {
	minioClient *minio.Client
	bucketName  string
	endpoint    string
}

// AdminUploadConfig holds MinIO configuration for admin uploads.
type AdminUploadConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// NewAdminUploadHandler creates a new upload handler.
func NewAdminUploadHandler(cfg AdminUploadConfig) (*AdminUploadHandler, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("creating minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("checking bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("creating bucket: %w", err)
		}
	}

	return &AdminUploadHandler{
		minioClient: client,
		bucketName:  cfg.Bucket,
		endpoint:    cfg.Endpoint,
	}, nil
}

// Upload handles POST /v1/admin/uploads — upload admin assets.
func (h *AdminUploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max for logos
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "file too large or invalid form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "MISSING_FILE", "file field is required")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]string{
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".svg":  "image/svg+xml",
		".webp": "image/webp",
		".ico":  "image/x-icon",
	}
	contentType, allowed := allowedExts[ext]
	if !allowed {
		writeError(w, http.StatusBadRequest, "INVALID_FILE_TYPE", "only PNG, JPG, SVG, WebP, ICO files are allowed")
		return
	}

	category := r.FormValue("category")
	if category == "" {
		category = "branding"
	}

	objectKey := fmt.Sprintf("admin/%s/%s%s", category, uuid.New().String(), ext)

	_, err = h.minioClient.PutObject(r.Context(), h.bucketName, objectKey, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "UPLOAD_FAILED", "failed to upload file")
		return
	}

	fileURL := fmt.Sprintf("http://%s/%s/%s", h.endpoint, h.bucketName, objectKey)

	writeJSON(w, http.StatusOK, envelope{
		"data": map[string]string{
			"url":      fileURL,
			"key":      objectKey,
			"filename": header.Filename,
		},
	})
}
