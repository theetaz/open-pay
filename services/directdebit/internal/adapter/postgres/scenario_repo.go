package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/directdebit/internal/domain"
)

// ScenarioRepository implements scenario code persistence.
type ScenarioRepository struct {
	pool *pgxpool.Pool
}

// NewScenarioRepository creates a new scenario repository.
func NewScenarioRepository(pool *pgxpool.Pool) *ScenarioRepository {
	return &ScenarioRepository{pool: pool}
}

func (r *ScenarioRepository) Create(ctx context.Context, s *domain.ScenarioCode) error {
	query := `INSERT INTO scenario_codes (id, scenario_id, scenario_name, payment_provider, max_limit, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		s.ID, s.ScenarioID, s.ScenarioName, s.PaymentProvider, s.MaxLimit, s.IsActive, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting scenario code: %w", err)
	}
	return nil
}

func (r *ScenarioRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ScenarioCode, error) {
	query := `SELECT id, scenario_id, scenario_name, payment_provider, max_limit, is_active, created_at, updated_at
		FROM scenario_codes WHERE id = $1`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("querying scenario: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrScenarioNotFound
	}
	return scanScenario(rows)
}

func (r *ScenarioRepository) List(ctx context.Context, provider string, activeOnly bool) ([]*domain.ScenarioCode, error) {
	query := `SELECT id, scenario_id, scenario_name, payment_provider, max_limit, is_active, created_at, updated_at
		FROM scenario_codes WHERE 1=1`
	args := []any{}
	argIdx := 1

	if provider != "" {
		query += fmt.Sprintf(" AND payment_provider = $%d", argIdx)
		args = append(args, provider)
		argIdx++
	}
	if activeOnly {
		query += fmt.Sprintf(" AND is_active = $%d", argIdx)
		args = append(args, true)
	}

	query += " ORDER BY payment_provider, scenario_name"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing scenarios: %w", err)
	}
	defer rows.Close()

	var scenarios []*domain.ScenarioCode
	for rows.Next() {
		s, err := scanScenario(rows)
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, s)
	}
	return scenarios, nil
}

func scanScenario(rows pgx.Rows) (*domain.ScenarioCode, error) {
	var s domain.ScenarioCode
	err := rows.Scan(
		&s.ID, &s.ScenarioID, &s.ScenarioName, &s.PaymentProvider, &s.MaxLimit, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning scenario: %w", err)
	}
	return &s, nil
}
