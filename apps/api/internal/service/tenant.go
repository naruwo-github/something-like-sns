package service

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	v1 "github.com/example/something-like-sns/apps/api/gen/sns/v1"
	v1connect "github.com/example/something-like-sns/apps/api/gen/sns/v1/v1connect"
)

type TenantServiceServer struct {
    db               *sql.DB
    allowDevHeaders  bool
}

func NewTenantServiceServer(db *sql.DB, allowDevHeaders bool) *TenantServiceServer {
    return &TenantServiceServer{db: db, allowDevHeaders: allowDevHeaders}
}

// MountHandler returns an http.Handler for this service
func (s *TenantServiceServer) MountHandler() (string, http.Handler) {
    path, handler := v1connect.NewTenantServiceHandler(s)
    return path, handler
}

func (s *TenantServiceServer) ResolveTenant(ctx context.Context, req *connect.Request[v1.ResolveTenantRequest]) (*connect.Response[v1.ResolveTenantResponse], error) {
    host := strings.TrimSpace(req.Msg.GetHost())
    if host == "" {
        return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("host is required"))
    }

    var tenantID uint64
    var slug string
    // Try exact domain match first
    err := s.db.QueryRowContext(ctx, "SELECT t.id, t.slug FROM tenant_domains d JOIN tenants t ON t.id=d.tenant_id WHERE d.domain=?", host).Scan(&tenantID, &slug)
    if err == sql.ErrNoRows {
        // Fallback: subdomain slug before first dot
        if idx := strings.IndexByte(host, '.'); idx > 0 {
            guess := host[:idx]
            err = s.db.QueryRowContext(ctx, "SELECT id, slug FROM tenants WHERE slug=?", guess).Scan(&tenantID, &slug)
        }
    }
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, connect.NewError(connect.CodeNotFound, errors.New("tenant not found"))
        }
        return nil, connect.NewError(connect.CodeInternal, err)
    }

    res := connect.NewResponse(&v1.ResolveTenantResponse{TenantId: tenantID, Slug: slug})
    return res, nil
}

func (s *TenantServiceServer) GetMe(ctx context.Context, req *connect.Request[v1.GetMeRequest]) (*connect.Response[v1.GetMeResponse], error) {
    if !s.allowDevHeaders {
        return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("dev headers disabled"))
    }
    tenantSlug := req.Header().Get("X-Tenant")
    authSub := req.Header().Get("X-User")
    if tenantSlug == "" || authSub == "" {
        return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing X-Tenant or X-User"))
    }

    // Resolve tenant ID
    var tenantID uint64
    if err := s.db.QueryRowContext(ctx, "SELECT id FROM tenants WHERE slug=?", tenantSlug).Scan(&tenantID); err != nil {
        if err == sql.ErrNoRows {
            return nil, connect.NewError(connect.CodeNotFound, errors.New("tenant not found"))
        }
        return nil, connect.NewError(connect.CodeInternal, err)
    }

    // Upsert user
    _, _ = s.db.ExecContext(ctx, "INSERT INTO users (auth_sub, display_name) VALUES (?, ?) ON DUPLICATE KEY UPDATE display_name=VALUES(display_name)", authSub, authSub)
    var userID uint64
    if err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE auth_sub=?", authSub).Scan(&userID); err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }

    // Ensure membership exists as member
    _, _ = s.db.ExecContext(ctx, "INSERT INTO tenant_memberships (tenant_id, user_id, role) VALUES (?, ?, 'member') ON DUPLICATE KEY UPDATE role=role", tenantID, userID)

    // Collect memberships
    rows, err := s.db.QueryContext(ctx, "SELECT m.tenant_id, m.role, t.slug FROM tenant_memberships m JOIN tenants t ON t.id=m.tenant_id WHERE m.user_id=? ORDER BY m.tenant_id", userID)
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    defer rows.Close()

    memberships := make([]*v1.TenantMembership, 0, 4)
    for rows.Next() {
        var m v1.TenantMembership
        if err := rows.Scan(&m.TenantId, &m.Role, &m.TenantSlug); err != nil {
            return nil, connect.NewError(connect.CodeInternal, err)
        }
        memberships = append(memberships, &m)
    }
    if err := rows.Err(); err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }

    res := connect.NewResponse(&v1.GetMeResponse{
        UserId:      userID,
        DisplayName: authSub,
        Memberships: memberships,
    })
    return res, nil
}
