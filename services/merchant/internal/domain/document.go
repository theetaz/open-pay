package domain

import "github.com/google/uuid"

// Document represents uploaded document metadata.
type Document struct {
	ID          uuid.UUID
	MerchantID  uuid.UUID
	Category    string
	Filename    string
	ObjectKey   string
	ContentType string
	FileSize    int64
}
