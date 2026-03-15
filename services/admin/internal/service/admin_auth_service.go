package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/openlankapay/openlankapay/services/admin/internal/domain"
)

type AdminUserRepository interface {
	GetByEmail(ctx context.Context, email string) (*domain.AdminUser, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AdminUser, error)
	Create(ctx context.Context, user *domain.AdminUser) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	GetRoleByName(ctx context.Context, name string) (*domain.AdminRole, error)
}

type AdminLoginResult struct {
	User         *domain.AdminUser
	AccessToken  string
	RefreshToken string
}

type AdminAuthService struct {
	repo       AdminUserRepository
	jwtSecret  string
	tokenTTL   time.Duration
	refreshTTL time.Duration
}

func NewAdminAuthService(repo AdminUserRepository, jwtSecret string) *AdminAuthService {
	return &AdminAuthService{
		repo:       repo,
		jwtSecret:  jwtSecret,
		tokenTTL:   8 * time.Hour,
		refreshTTL: 24 * time.Hour,
	}
}

func (s *AdminAuthService) Login(ctx context.Context, email, password string) (*AdminLoginResult, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, domain.ErrAdminAccountInactive
	}

	if !user.VerifyPassword(password) {
		return nil, domain.ErrInvalidCredentials
	}

	_ = s.repo.UpdateLastLogin(ctx, user.ID)

	accessToken, err := auth.GenerateToken(user.ID, uuid.Nil, "PLATFORM_"+user.Role.Name, nil, s.jwtSecret, s.tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, s.jwtSecret, s.refreshTTL)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	return &AdminLoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AdminAuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*domain.AdminUser, error) {
	return s.repo.GetByID(ctx, userID)
}

func (s *AdminAuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*AdminLoginResult, error) {
	userID, err := auth.ValidateRefreshToken(refreshTokenStr, s.jwtSecret)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, domain.ErrAdminAccountInactive
	}

	accessToken, err := auth.GenerateToken(user.ID, uuid.Nil, "PLATFORM_"+user.Role.Name, nil, s.jwtSecret, s.tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	newRefreshToken, err := auth.GenerateRefreshToken(user.ID, s.jwtSecret, s.refreshTTL)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	return &AdminLoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
