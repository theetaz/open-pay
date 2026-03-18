package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
)

// DirectorRepository manages director records in PostgreSQL.
type DirectorRepository struct {
	pool *pgxpool.Pool
}

// NewDirectorRepository creates a new DirectorRepository.
func NewDirectorRepository(pool *pgxpool.Pool) *DirectorRepository {
	return &DirectorRepository{pool: pool}
}

// Create inserts a new director record.
func (r *DirectorRepository) Create(ctx context.Context, d *domain.Director) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO merchant_directors (id, merchant_id, email, verification_token, token_expires_at, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		d.ID, d.MerchantID, d.Email, d.VerificationToken, d.TokenExpiresAt, d.Status, d.CreatedAt, d.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrDuplicateDirector
		}
		return fmt.Errorf("inserting director: %w", err)
	}
	return nil
}

// GetByID retrieves a director by ID.
func (r *DirectorRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Director, error) {
	return r.getOne(ctx, "id = $1", id)
}

// GetByToken retrieves a director by verification token.
func (r *DirectorRepository) GetByToken(ctx context.Context, token string) (*domain.Director, error) {
	return r.getOne(ctx, "verification_token = $1", token)
}

// ListByMerchant returns all directors for a merchant.
func (r *DirectorRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.Director, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, merchant_id, email, COALESCE(full_name,''), date_of_birth,
			COALESCE(nic_passport_number,''), COALESCE(phone,''), COALESCE(address,''),
			COALESCE(document_object_key,''), COALESCE(document_filename,''),
			verification_token, token_expires_at, status, consented_at, verified_at, created_at, updated_at
		FROM merchant_directors
		WHERE merchant_id = $1
		ORDER BY created_at ASC`, merchantID)
	if err != nil {
		return nil, fmt.Errorf("listing directors: %w", err)
	}
	defer rows.Close()

	var directors []*domain.Director
	for rows.Next() {
		d, err := scanDirector(rows)
		if err != nil {
			return nil, err
		}
		directors = append(directors, d)
	}
	return directors, nil
}

// CountByMerchant returns the number of directors for a merchant.
func (r *DirectorRepository) CountByMerchant(ctx context.Context, merchantID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM merchant_directors WHERE merchant_id = $1`, merchantID).Scan(&count)
	return count, err
}

// Update updates a director record.
func (r *DirectorRepository) Update(ctx context.Context, d *domain.Director) error {
	d.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE merchant_directors SET
			email = $2, full_name = $3, date_of_birth = $4, nic_passport_number = $5,
			phone = $6, address = $7, document_object_key = $8, document_filename = $9,
			verification_token = $10, token_expires_at = $11, status = $12,
			consented_at = $13, verified_at = $14, updated_at = $15
		WHERE id = $1`,
		d.ID, d.Email, d.FullName, d.DateOfBirth, d.NICPassportNumber,
		d.Phone, d.Address, d.DocumentObjectKey, d.DocumentFilename,
		d.VerificationToken, d.TokenExpiresAt, d.Status,
		d.ConsentedAt, d.VerifiedAt, d.UpdatedAt,
	)
	return err
}

// Delete removes a director record.
func (r *DirectorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM merchant_directors WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrDirectorNotFound
	}
	return nil
}

func (r *DirectorRepository) getOne(ctx context.Context, where string, args ...any) (*domain.Director, error) {
	row := r.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT id, merchant_id, email, COALESCE(full_name,''), date_of_birth,
			COALESCE(nic_passport_number,''), COALESCE(phone,''), COALESCE(address,''),
			COALESCE(document_object_key,''), COALESCE(document_filename,''),
			verification_token, token_expires_at, status, consented_at, verified_at, created_at, updated_at
		FROM merchant_directors
		WHERE %s`, where), args...)

	d := &domain.Director{}
	err := row.Scan(
		&d.ID, &d.MerchantID, &d.Email, &d.FullName, &d.DateOfBirth,
		&d.NICPassportNumber, &d.Phone, &d.Address,
		&d.DocumentObjectKey, &d.DocumentFilename,
		&d.VerificationToken, &d.TokenExpiresAt, &d.Status,
		&d.ConsentedAt, &d.VerifiedAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrDirectorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scanning director: %w", err)
	}
	return d, nil
}

type directorScanner interface {
	Scan(dest ...any) error
}

func scanDirector(s directorScanner) (*domain.Director, error) {
	d := &domain.Director{}
	err := s.Scan(
		&d.ID, &d.MerchantID, &d.Email, &d.FullName, &d.DateOfBirth,
		&d.NICPassportNumber, &d.Phone, &d.Address,
		&d.DocumentObjectKey, &d.DocumentFilename,
		&d.VerificationToken, &d.TokenExpiresAt, &d.Status,
		&d.ConsentedAt, &d.VerifiedAt, &d.CreatedAt, &d.UpdatedAt,
	)
	return d, err
}
