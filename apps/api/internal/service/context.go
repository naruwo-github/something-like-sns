package service

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"connectrpc.com/connect"
)

type RequestScope struct {
    TenantID uint64
    UserID   uint64
}

func resolveScopeFromDevHeaders(ctx context.Context, db *sql.DB, headers http.Header) (*RequestScope, error) {
    tenantSlug := headers.Get("X-Tenant")
    authSub := headers.Get("X-User")
    if tenantSlug == "" || authSub == "" {
        return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing X-Tenant or X-User"))
    }
    var tenantID uint64
    if err := db.QueryRowContext(ctx, "SELECT id FROM tenants WHERE slug=?", tenantSlug).Scan(&tenantID); err != nil {
        if err == sql.ErrNoRows {
            return nil, connect.NewError(connect.CodeNotFound, errors.New("tenant not found"))
        }
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    // Upsert user
    if _, err := db.ExecContext(ctx, "INSERT INTO users (auth_sub, display_name) VALUES (?, ?) ON DUPLICATE KEY UPDATE display_name=VALUES(display_name)", authSub, authSub); err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    var userID uint64
    if err := db.QueryRowContext(ctx, "SELECT id FROM users WHERE auth_sub=?", authSub).Scan(&userID); err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    // Ensure membership exists
    if _, err := db.ExecContext(ctx, "INSERT INTO tenant_memberships (tenant_id, user_id, role) VALUES (?, ?, 'member') ON DUPLICATE KEY UPDATE role=role", tenantID, userID); err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    return &RequestScope{TenantID: tenantID, UserID: userID}, nil
}


