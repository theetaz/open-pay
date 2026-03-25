package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
	"github.com/openlankapay/openlankapay/services/merchant/internal/service"
)

// CreateDirector handles POST /v1/merchants/{id}/directors.
func (h *MerchantHandler) CreateDirector(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant ID")
		return
	}

	if claims.MerchantID != id {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "cannot access another merchant")
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "email is required")
		return
	}

	director, err := h.svc.CreateDirector(r.Context(), id, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidDirector) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		if errors.Is(err, domain.ErrMaxDirectors) {
			writeError(w, http.StatusUnprocessableEntity, "MAX_DIRECTORS", err.Error())
			return
		}
		if errors.Is(err, domain.ErrDuplicateDirector) {
			writeError(w, http.StatusConflict, "DUPLICATE_DIRECTOR", "director email already exists for this merchant")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create director")
		return
	}

	writeJSON(w, http.StatusCreated, envelope{"data": directorResponse(director)})
}

// ListDirectors handles GET /v1/merchants/{id}/directors.
func (h *MerchantHandler) ListDirectors(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant ID")
		return
	}

	// Admin JWTs have zero-value MerchantID; merchant users must own the resource
	if claims.MerchantID != uuid.Nil && claims.MerchantID != id {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "cannot access another merchant")
		return
	}

	directors, err := h.svc.ListDirectors(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list directors")
		return
	}

	items := make([]map[string]any, 0, len(directors))
	for _, d := range directors {
		items = append(items, directorResponse(d))
	}

	writeJSON(w, http.StatusOK, envelope{"data": items})
}

// ResendDirectorVerification handles POST /v1/merchants/{id}/directors/{directorId}/resend.
func (h *MerchantHandler) ResendDirectorVerification(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant ID")
		return
	}

	if claims.MerchantID != id {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "cannot access another merchant")
		return
	}

	directorID, err := uuid.Parse(chi.URLParam(r, "directorId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid director ID")
		return
	}

	if err := h.svc.ResendDirectorVerification(r.Context(), id, directorID); err != nil {
		if errors.Is(err, domain.ErrDirectorNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "director not found")
			return
		}
		if errors.Is(err, domain.ErrInvalidDirector) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to resend verification")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "verification_sent"}})
}

// RemoveDirector handles DELETE /v1/merchants/{id}/directors/{directorId}.
func (h *MerchantHandler) RemoveDirector(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant ID")
		return
	}

	if claims.MerchantID != id {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "cannot access another merchant")
		return
	}

	directorID, err := uuid.Parse(chi.URLParam(r, "directorId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid director ID")
		return
	}

	if err := h.svc.RemoveDirector(r.Context(), id, directorID); err != nil {
		if errors.Is(err, domain.ErrDirectorNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "director not found")
			return
		}
		if errors.Is(err, domain.ErrInvalidDirector) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to remove director")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "removed"}})
}

// GetDirectorByToken handles GET /v1/public/directors/verify/{token}.
func (h *MerchantHandler) GetDirectorByToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "INVALID_TOKEN", "token is required")
		return
	}

	director, merchant, err := h.svc.GetDirectorByToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, domain.ErrDirectorNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "director not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get director")
		return
	}

	writeJSON(w, http.StatusOK, envelope{
		"data": map[string]any{
			"director":     directorResponse(director),
			"businessName": merchant.BusinessName,
		},
	})
}

// SubmitDirectorVerification handles POST /v1/public/directors/verify/{token}.
func (h *MerchantHandler) SubmitDirectorVerification(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "INVALID_TOKEN", "token is required")
		return
	}

	// Parse multipart form (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid multipart form")
		return
	}

	fullName := r.FormValue("fullName")
	dateOfBirthStr := r.FormValue("dateOfBirth")
	nicPassportNumber := r.FormValue("nicPassportNumber")
	phone := r.FormValue("phone")
	address := r.FormValue("address")
	consent := r.FormValue("consent")

	if fullName == "" || dateOfBirthStr == "" || nicPassportNumber == "" || phone == "" || address == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "fullName, dateOfBirth, nicPassportNumber, phone, and address are required")
		return
	}
	if consent != "true" {
		writeError(w, http.StatusBadRequest, "CONSENT_REQUIRED", "consent is required")
		return
	}

	dob, err := time.Parse("2006-01-02", dateOfBirthStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "dateOfBirth must be in YYYY-MM-DD format")
		return
	}

	file, header, err := r.FormFile("document")
	if err != nil {
		writeError(w, http.StatusBadRequest, "MISSING_FILE", "document file is required")
		return
	}
	defer func() { _ = file.Close() }()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]string{
		".pdf":  "application/pdf",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
	}
	contentType, allowed := allowedExts[ext]
	if !allowed {
		writeError(w, http.StatusBadRequest, "INVALID_FILE_TYPE", "only PDF, PNG, JPG files are allowed")
		return
	}

	// Look up the director first to get the ID for the object key
	director, _, err := h.svc.GetDirectorByToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, domain.ErrDirectorNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "director not found or token invalid")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to look up director")
		return
	}

	var objectKey string
	var documentFilename string

	if h.minioClient != nil {
		objectKey = fmt.Sprintf("directors/%s/%s%s", director.ID.String(), uuid.New().String(), ext)
		_, err = h.minioClient.PutObject(r.Context(), h.minioBucket, objectKey, file, header.Size, minio.PutObjectOptions{
			ContentType: contentType,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "UPLOAD_FAILED", "failed to upload document")
			return
		}
		documentFilename = header.Filename
	}

	dobPtr := &dob
	updated, err := h.svc.SubmitDirectorVerification(r.Context(), token, service.SubmitDirectorInput{
		FullName:          fullName,
		DateOfBirth:       dobPtr,
		NICPassportNumber: nicPassportNumber,
		Phone:             phone,
		Address:           address,
		DocumentObjectKey: objectKey,
		DocumentFilename:  documentFilename,
	})
	if err != nil {
		if errors.Is(err, domain.ErrTokenExpired) {
			writeError(w, http.StatusGone, "TOKEN_EXPIRED", "verification token has expired")
			return
		}
		if errors.Is(err, domain.ErrInvalidDirector) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to submit verification")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": directorResponse(updated)})
}

func directorResponse(d *domain.Director) map[string]any {
	resp := map[string]any{
		"id":                d.ID.String(),
		"merchantId":        d.MerchantID.String(),
		"email":             d.Email,
		"fullName":          d.FullName,
		"nicPassportNumber": d.NICPassportNumber,
		"phone":             d.Phone,
		"address":           d.Address,
		"documentObjectKey": d.DocumentObjectKey,
		"documentFilename":  d.DocumentFilename,
		"status":            d.Status,
		"tokenExpiresAt":    d.TokenExpiresAt.Format(time.RFC3339),
		"tokenExpired":      d.IsTokenExpired(),
		"createdAt":         d.CreatedAt.Format(time.RFC3339),
		"updatedAt":         d.UpdatedAt.Format(time.RFC3339),
	}
	if d.DateOfBirth != nil {
		resp["dateOfBirth"] = d.DateOfBirth.Format("2006-01-02")
	} else {
		resp["dateOfBirth"] = nil
	}
	if d.ConsentedAt != nil {
		resp["consentedAt"] = d.ConsentedAt.Format(time.RFC3339)
	} else {
		resp["consentedAt"] = nil
	}
	if d.VerifiedAt != nil {
		resp["verifiedAt"] = d.VerifiedAt.Format(time.RFC3339)
	} else {
		resp["verifiedAt"] = nil
	}
	return resp
}
