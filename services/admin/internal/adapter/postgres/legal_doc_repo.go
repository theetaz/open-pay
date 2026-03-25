package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LegalDocument represents a versioned legal document.
type LegalDocument struct {
	ID           uuid.UUID
	Type         string
	Version      int
	Title        string
	Content      string
	IsActive     bool
	CreatedBy    *uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	PdfObjectKey string
}

// LegalDocumentRepository provides access to legal documents.
type LegalDocumentRepository struct {
	pool *pgxpool.Pool
}

// NewLegalDocumentRepository creates a new repo.
func NewLegalDocumentRepository(pool *pgxpool.Pool) *LegalDocumentRepository {
	return &LegalDocumentRepository{pool: pool}
}

// GetActiveByType returns the currently active document of a given type.
func (r *LegalDocumentRepository) GetActiveByType(ctx context.Context, docType string) (*LegalDocument, error) {
	doc := &LegalDocument{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, type, version, title, content, is_active, created_by, created_at, updated_at, COALESCE(pdf_object_key, '')
		 FROM legal_documents WHERE type = $1 AND is_active = TRUE LIMIT 1`, docType,
	).Scan(&doc.ID, &doc.Type, &doc.Version, &doc.Title, &doc.Content, &doc.IsActive, &doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt, &doc.PdfObjectKey)
	if err != nil {
		return nil, fmt.Errorf("legal document not found: %w", err)
	}
	return doc, nil
}

// List returns all legal documents ordered by type and version.
func (r *LegalDocumentRepository) List(ctx context.Context) ([]*LegalDocument, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, type, version, title, content, is_active, created_by, created_at, updated_at, COALESCE(pdf_object_key, '')
		 FROM legal_documents ORDER BY type, version DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*LegalDocument
	for rows.Next() {
		d := &LegalDocument{}
		if err := rows.Scan(&d.ID, &d.Type, &d.Version, &d.Title, &d.Content, &d.IsActive, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt, &d.PdfObjectKey); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

// Create inserts a new legal document version.
func (r *LegalDocumentRepository) Create(ctx context.Context, doc *LegalDocument) error {
	doc.ID = uuid.New()
	doc.CreatedAt = time.Now().UTC()
	doc.UpdatedAt = doc.CreatedAt

	_, err := r.pool.Exec(ctx,
		`INSERT INTO legal_documents (id, type, version, title, content, is_active, created_by, created_at, updated_at, pdf_object_key)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		doc.ID, doc.Type, doc.Version, doc.Title, doc.Content, doc.IsActive, doc.CreatedBy, doc.CreatedAt, doc.UpdatedAt, doc.PdfObjectKey,
	)
	return err
}

// Activate sets a document as active and deactivates all other versions of the same type.
func (r *LegalDocumentRepository) Activate(ctx context.Context, id uuid.UUID) error {
	// Get the document type
	var docType string
	err := r.pool.QueryRow(ctx, `SELECT type FROM legal_documents WHERE id = $1`, id).Scan(&docType)
	if err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	// Deactivate all of the same type
	_, err = r.pool.Exec(ctx, `UPDATE legal_documents SET is_active = FALSE, updated_at = NOW() WHERE type = $1`, docType)
	if err != nil {
		return err
	}

	// Activate the specified one
	_, err = r.pool.Exec(ctx, `UPDATE legal_documents SET is_active = TRUE, updated_at = NOW() WHERE id = $1`, id)
	return err
}

// Update updates a legal document's content.
func (r *LegalDocumentRepository) Update(ctx context.Context, doc *LegalDocument) error {
	doc.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`UPDATE legal_documents SET title = $1, content = $2, updated_at = $3, pdf_object_key = $4 WHERE id = $5`,
		doc.Title, doc.Content, doc.UpdatedAt, doc.PdfObjectKey, doc.ID,
	)
	return err
}
