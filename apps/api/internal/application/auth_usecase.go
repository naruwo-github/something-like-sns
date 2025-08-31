package application

import (
	"context"
	"errors"
	"strings"

	"github.com/example/something-like-sns/apps/api/internal/domain"
	"github.com/example/something-like-sns/apps/api/internal/port"
)

type authUsecase struct {
	store port.Store
}

func NewAuthUsecase(store port.Store) port.AuthUsecase {
	return &authUsecase{store: store}
}

func (u *authUsecase) ResolveScope(ctx context.Context, tenantSlug, userAuthSub string) (*domain.Scope, error) {
	if tenantSlug == "" || userAuthSub == "" {
		return nil, errors.New("missing tenant slug or user auth sub")
	}
	tenant, err := u.store.AuthRepository().FindTenantBySlug(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}

	userID, err := u.store.AuthRepository().FindOrCreateUser(ctx, userAuthSub, userAuthSub)
	if err != nil {
		return nil, err
	}

	if err := u.store.AuthRepository().EnsureMembership(ctx, tenant.ID, userID, "member"); err != nil {
		return nil, err
	}

	return &domain.Scope{TenantID: tenant.ID, UserID: userID}, nil
}

func (u *authUsecase) ResolveTenant(ctx context.Context, host string) (*domain.Tenant, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, errors.New("host is required")
	}
	return u.store.AuthRepository().FindTenantByHost(ctx, host)
}

func (u *authUsecase) GetMe(ctx context.Context, userID uint64) (*domain.User, error) {
	user, err := u.store.AuthRepository().FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	memberships, err := u.store.AuthRepository().FindUserMemberships(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Memberships = memberships
	return user, nil
}
