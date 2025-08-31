package mysql

import (
	"context"
	"database/sql"
	"strings"

	"github.com/example/something-like-sns/apps/api/internal/domain"
)

type authRepository struct {
	q DBTX
}

func (r *authRepository) FindTenantByHost(ctx context.Context, host string) (*domain.Tenant, error) {
	var t domain.Tenant
	err := r.q.QueryRowContext(ctx, "SELECT t.id, t.slug FROM tenant_domains d JOIN tenants t ON t.id=d.tenant_id WHERE d.domain=?", host).Scan(&t.ID, &t.Slug)
	if err == sql.ErrNoRows {
		if idx := strings.IndexByte(host, '.'); idx > 0 {
			guess := host[:idx]
			err = r.q.QueryRowContext(ctx, "SELECT id, slug FROM tenants WHERE slug=?", guess).Scan(&t.ID, &t.Slug)
		}
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *authRepository) FindTenantBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	var t domain.Tenant
	err := r.q.QueryRowContext(ctx, "SELECT id, slug FROM tenants WHERE slug=?", slug).Scan(&t.ID, &t.Slug)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *authRepository) FindOrCreateUser(ctx context.Context, authSub, displayName string) (uint64, error) {
	_, err := r.q.ExecContext(ctx, "INSERT INTO users (auth_sub, display_name) VALUES (?, ?) ON DUPLICATE KEY UPDATE display_name=VALUES(display_name)", authSub, displayName)
	if err != nil {
		return 0, err
	}
	var userID uint64
	if err := r.q.QueryRowContext(ctx, "SELECT id FROM users WHERE auth_sub=?", authSub).Scan(&userID); err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *authRepository) FindUserByID(ctx context.Context, userID uint64) (*domain.User, error) {
	var u domain.User
	err := r.q.QueryRowContext(ctx, "SELECT id, display_name FROM users WHERE id=?", userID).Scan(&u.ID, &u.DisplayName)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *authRepository) EnsureMembership(ctx context.Context, tenantID, userID uint64, role string) error {
	_, err := r.q.ExecContext(ctx, "INSERT INTO tenant_memberships (tenant_id, user_id, role) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE role=role", tenantID, userID, role)
	return err
}

func (r *authRepository) FindUserMemberships(ctx context.Context, userID uint64) ([]*domain.TenantMembership, error) {
	rows, err := r.q.QueryContext(ctx, "SELECT m.tenant_id, m.role, t.slug FROM tenant_memberships m JOIN tenants t ON t.id=m.tenant_id WHERE m.user_id=? ORDER BY m.tenant_id", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memberships := make([]*domain.TenantMembership, 0, 4)
	for rows.Next() {
		var m domain.TenantMembership
		if err := rows.Scan(&m.TenantID, &m.Role, &m.TenantSlug); err != nil {
			return nil, err
		}
		memberships = append(memberships, &m)
	}
	return memberships, rows.Err()
}