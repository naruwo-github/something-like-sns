package rpc

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	v1 "github.com/example/something-like-sns/apps/api/gen/sns/v1"
	"github.com/example/something-like-sns/apps/api/gen/sns/v1/v1connect"
	"github.com/example/something-like-sns/apps/api/internal/port"
)

type TenantHandler struct {
	authUsecase     port.AuthUsecase
	allowDevHeaders bool
}

func NewTenantHandler(au port.AuthUsecase, allowDev bool) *TenantHandler {
	return &TenantHandler{authUsecase: au, allowDevHeaders: allowDev}
}

func (s *TenantHandler) MountHandler() (string, http.Handler) {
	path, handler := v1connect.NewTenantServiceHandler(s)
	return path, handler
}

func (s *TenantHandler) ResolveTenant(ctx context.Context, req *connect.Request[v1.ResolveTenantRequest]) (*connect.Response[v1.ResolveTenantResponse], error) {
	tenant, err := s.authUsecase.ResolveTenant(ctx, req.Msg.GetHost())
	if err != nil {
		// TODO: Map domain errors to connect errors
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&v1.ResolveTenantResponse{TenantId: tenant.ID, Slug: tenant.Slug}), nil
}

func (s *TenantHandler) GetMe(ctx context.Context, req *connect.Request[v1.GetMeRequest]) (*connect.Response[v1.GetMeResponse], error) {
	if !s.allowDevHeaders {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("dev headers disabled"))
	}
	tenantSlug := req.Header().Get("X-Tenant")
	authSub := req.Header().Get("X-User")

	user, err := s.authUsecase.GetMe(ctx, tenantSlug, authSub)
	if err != nil {
		// TODO: Map domain errors to connect errors
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	memberships := make([]*v1.TenantMembership, len(user.Memberships))
	for i, m := range user.Memberships {
		memberships[i] = &v1.TenantMembership{
			TenantId:   m.TenantID,
			Role:       m.Role,
			TenantSlug: m.TenantSlug,
		}
	}

	return connect.NewResponse(&v1.GetMeResponse{
		UserId:      user.ID,
		DisplayName: user.DisplayName,
		Memberships: memberships,
	}), nil
}