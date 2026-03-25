package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
)

// DocumentRepository implements handler.DocumentRepository with PostgreSQL.
type DocumentRepository struct {
	pool *pgxpool.Pool
}

// NewDocumentRepository creates a new DocumentRepository.
func NewDocumentRepository(pool *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{pool: pool}
}

func (r *DocumentRepository) Create(ctx context.Context, doc *domain.Document) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO documents (id, merchant_id, category, filename, object_key, content_type, file_size)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		doc.ID, doc.MerchantID, doc.Category, doc.Filename, doc.ObjectKey, doc.ContentType, doc.FileSize,
	)
	return err
}

func (r *DocumentRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.Document, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, merchant_id, category, filename, object_key, content_type, file_size
		 FROM documents WHERE merchant_id = $1 ORDER BY uploaded_at DESC`,
		merchantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		d := &domain.Document{}
		if err := rows.Scan(&d.ID, &d.MerchantID, &d.Category, &d.Filename, &d.ObjectKey, &d.ContentType, &d.FileSize); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (r *DocumentRepository) DeleteByKey(ctx context.Context, merchantID uuid.UUID, objectKey string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM documents WHERE merchant_id = $1 AND object_key = $2`,
		merchantID, objectKey,
	)
	return err
}
