package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EmailTemplate represents a stored email template.
type EmailTemplate struct {
	ID        uuid.UUID
	EventType string
	Name      string
	Subject   string
	BodyHTML  string
	Variables []string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// EmailTemplateRepository provides access to email templates.
type EmailTemplateRepository struct {
	pool *pgxpool.Pool
}

// NewEmailTemplateRepository creates a new repo.
func NewEmailTemplateRepository(pool *pgxpool.Pool) *EmailTemplateRepository {
	return &EmailTemplateRepository{pool: pool}
}

// GetByEventType returns the active template for a given event type.
func (r *EmailTemplateRepository) GetByEventType(ctx context.Context, eventType string) (*EmailTemplate, error) {
	t := &EmailTemplate{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, event_type, name, subject, body_html, variables, is_active, created_at, updated_at
		 FROM email_templates WHERE event_type = $1 AND is_active = TRUE LIMIT 1`, eventType,
	).Scan(&t.ID, &t.EventType, &t.Name, &t.Subject, &t.BodyHTML, &t.Variables, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// List returns all email templates.
func (r *EmailTemplateRepository) List(ctx context.Context) ([]*EmailTemplate, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, event_type, name, subject, body_html, variables, is_active, created_at, updated_at
		 FROM email_templates ORDER BY event_type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []*EmailTemplate
	for rows.Next() {
		t := &EmailTemplate{}
		if err := rows.Scan(&t.ID, &t.EventType, &t.Name, &t.Subject, &t.BodyHTML, &t.Variables, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

// Create inserts a new email template.
func (r *EmailTemplateRepository) Create(ctx context.Context, t *EmailTemplate) error {
	t.ID = uuid.New()
	t.CreatedAt = time.Now().UTC()
	t.UpdatedAt = t.CreatedAt
	t.IsActive = true

	_, err := r.pool.Exec(ctx,
		`INSERT INTO email_templates (id, event_type, name, subject, body_html, variables, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		t.ID, t.EventType, t.Name, t.Subject, t.BodyHTML, t.Variables, t.IsActive, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

// Update updates an email template.
func (r *EmailTemplateRepository) Update(ctx context.Context, t *EmailTemplate) error {
	t.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`UPDATE email_templates SET name = $1, subject = $2, body_html = $3, variables = $4, updated_at = $5 WHERE id = $6`,
		t.Name, t.Subject, t.BodyHTML, t.Variables, t.UpdatedAt, t.ID,
	)
	return err
}
