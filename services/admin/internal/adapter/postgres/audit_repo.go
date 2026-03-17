package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/admin/internal/domain"
)

type AuditRepository struct {
	pool *pgxpool.Pool
}

func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

func (r *AuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	changesJSON, _ := json.Marshal(log.Changes)
	metadataJSON, _ := json.Marshal(log.Metadata)

	query := `INSERT INTO audit_logs (id, actor_id, actor_type, merchant_id, action, resource_type, resource_id,
		changes, metadata, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	// Handle empty IP address — INET type doesn't accept empty string
	var ipAddr any = log.IPAddress
	if log.IPAddress == "" {
		ipAddr = nil
	}

	_, err := r.pool.Exec(ctx, query,
		log.ID, log.ActorID, log.ActorType, log.MerchantID, log.Action, log.ResourceType, log.ResourceID,
		changesJSON, metadataJSON, ipAddr, log.UserAgent, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting audit log: %w", err)
	}
	return nil
}

func (r *AuditRepository) List(ctx context.Context, params ListParams) ([]*domain.AuditLog, int, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}

	conditions := []string{"1=1"}
	args := []any{}
	argIdx := 1

	if params.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIdx))
		args = append(args, params.Action)
		argIdx++
	}
	if params.ActorType != "" {
		conditions = append(conditions, fmt.Sprintf("actor_type = $%d", argIdx))
		args = append(args, params.ActorType)
		argIdx++
	}
	if params.ResourceType != "" {
		conditions = append(conditions, fmt.Sprintf("resource_type = $%d", argIdx))
		args = append(args, params.ResourceType)
		argIdx++
	}
	if params.MerchantID != nil {
		conditions = append(conditions, fmt.Sprintf("merchant_id = $%d", argIdx))
		args = append(args, *params.MerchantID)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	countQuery := "SELECT COUNT(*) FROM audit_logs WHERE " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting audit logs: %w", err)
	}

	offset := (params.Page - 1) * params.PerPage
	selectQuery := fmt.Sprintf(`SELECT id, actor_id, actor_type, merchant_id, action, resource_type, resource_id,
		changes, metadata, COALESCE(host(ip_address),''), user_agent, created_at
		FROM audit_logs WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1)
	args = append(args, params.PerPage, offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		l, err := scanAuditLog(rows)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	return logs, total, nil
}

type ListParams struct {
	Page         int
	PerPage      int
	Action       string
	ActorType    string
	ResourceType string
	MerchantID   *uuid.UUID
}

func scanAuditLog(rows pgx.Rows) (*domain.AuditLog, error) {
	var l domain.AuditLog
	var changesJSON, metadataJSON []byte

	err := rows.Scan(
		&l.ID, &l.ActorID, &l.ActorType, &l.MerchantID, &l.Action, &l.ResourceType, &l.ResourceID,
		&changesJSON, &metadataJSON, &l.IPAddress, &l.UserAgent, &l.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning audit log: %w", err)
	}

	_ = json.Unmarshal(changesJSON, &l.Changes)
	_ = json.Unmarshal(metadataJSON, &l.Metadata)
	return &l, nil
}

func (r *AuditRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error) {
	query := `SELECT id, actor_id, actor_type, merchant_id, action, resource_type, resource_id,
		changes, metadata, COALESCE(host(ip_address),''), user_agent, created_at
		FROM audit_logs WHERE id = $1`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("querying audit log: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrAuditNotFound
	}
	return scanAuditLog(rows)
}
