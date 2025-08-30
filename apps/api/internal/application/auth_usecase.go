package application

import (
	"context"
	"errors"
	"strings"

	"github.com/example/something-like-sns/apps/api/internal/domain"
	"github.com/example/something-like-sns/apps/api/internal/port"
)

type authUsecase struct {
	authRepo port.AuthRepository
}

func NewAuthUsecase(ar port.AuthRepository) port.AuthUsecase {
	return &authUsecase{authRepo: ar}
}

func (u *authUsecase) ResolveScope(ctx context.Context, tenantSlug, userAuthSub string) (*domain.Scope, error) {
	if tenantSlug == "" || userAuthSub == "" {
		return nil, errors.New("missing tenant slug or user auth sub")
	}
	tenant, err := u.authRepo.FindTenantBySlug(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}

	userID, err := u.authRepo.FindOrCreateUser(ctx, userAuthSub, userAuthSub)
	if err != nil {
		return nil, err
	}

	if err := u.authRepo.EnsureMembership(ctx, tenant.ID, userID, "member"); err != nil {
		return nil, err
	}

	return &domain.Scope{TenantID: tenant.ID, UserID: userID}, nil
}

func (u *authUsecase) ResolveTenant(ctx context.Context, host string) (*domain.Tenant, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, errors.New("host is required")
	}
	return u.authRepo.FindTenantByHost(ctx, host)
}

func (u *authUsecase) GetMe(ctx context.Context, tenantSlug, userAuthSub string) (*domain.User, error) {
	scope, err := u.ResolveScope(ctx, tenantSlug, userAuthSub)
	if err != nil {
		return nil, err
	}

	memberships, err := u.authRepo.FindUserMemberships(ctx, scope.UserID)
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:          scope.UserID,
		DisplayName: userAuthSub,
		Memberships: memberships,
	}, nil
}
