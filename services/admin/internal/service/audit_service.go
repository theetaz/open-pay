package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/admin/internal/adapter/postgres"
	"github.com/openlankapay/openlankapay/services/admin/internal/domain"
)

type AuditRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error)
	List(ctx context.Context, params postgres.ListParams) ([]*domain.AuditLog, int, error)
}

type AuditService struct {
	repo AuditRepository
}

func NewAuditService(repo AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) CreateLog(ctx context.Context, input domain.AuditInput) (*domain.AuditLog, error) {
	log, err := domain.NewAuditLog(input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, log); err != nil {
		return nil, fmt.Errorf("storing audit log: %w", err)
	}
	return log, nil
}

func (s *AuditService) GetByID(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AuditService) List(ctx context.Context, params postgres.ListParams) ([]*domain.AuditLog, int, error) {
	return s.repo.List(ctx, params)
}
