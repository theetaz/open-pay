package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PlatformSetting represents a key-value setting.
type PlatformSetting struct {
	Key         string
	Value       string
	Description string
	Category    string
	UpdatedBy   *uuid.UUID
	UpdatedAt   time.Time
}

// SettingsRepository provides access to platform settings.
type SettingsRepository struct {
	pool *pgxpool.Pool
}

// NewSettingsRepository creates a new settings repo.
func NewSettingsRepository(pool *pgxpool.Pool) *SettingsRepository {
	return &SettingsRepository{pool: pool}
}

// GetAll returns all settings.
func (r *SettingsRepository) GetAll(ctx context.Context) ([]*PlatformSetting, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT key, value, description, category, updated_by, updated_at FROM platform_settings ORDER BY category, key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*PlatformSetting
	for rows.Next() {
		s := &PlatformSetting{}
		if err := rows.Scan(&s.Key, &s.Value, &s.Description, &s.Category, &s.UpdatedBy, &s.UpdatedAt); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	return settings, rows.Err()
}

// GetByCategory returns settings for a specific category.
func (r *SettingsRepository) GetByCategory(ctx context.Context, category string) ([]*PlatformSetting, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT key, value, description, category, updated_by, updated_at FROM platform_settings WHERE category = $1 ORDER BY key`, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*PlatformSetting
	for rows.Next() {
		s := &PlatformSetting{}
		if err := rows.Scan(&s.Key, &s.Value, &s.Description, &s.Category, &s.UpdatedBy, &s.UpdatedAt); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	return settings, rows.Err()
}

// Update updates a single setting.
func (r *SettingsRepository) Update(ctx context.Context, key, value string, updatedBy uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE platform_settings SET value = $1, updated_by = $2, updated_at = NOW() WHERE key = $3`,
		value, updatedBy, key)
	return err
}

// BulkUpdate updates multiple settings at once.
func (r *SettingsRepository) BulkUpdate(ctx context.Context, updates map[string]string, updatedBy uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for key, value := range updates {
		_, err := tx.Exec(ctx,
			`UPDATE platform_settings SET value = $1, updated_by = $2, updated_at = NOW() WHERE key = $3`,
			value, updatedBy, key)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
