package rpc

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/example/something-like-sns/apps/api/internal/domain"
	"github.com/example/something-like-sns/apps/api/internal/port"
)

type contextKey string

const scopeContextKey = contextKey("scope")

// NewAuthInterceptor creates a new connect.Interceptor for handling authentication.
func NewAuthInterceptor(authUsecase port.AuthUsecase, allowDevHeaders bool) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Bypass auth for TenantService.ResolveTenant as it's used for public tenant resolution.
			if req.Spec().Procedure == "/sns.v1.TenantService/ResolveTenant" {
				return next(ctx, req)
			}

			if !allowDevHeaders {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("dev headers disabled"))
			}

			tenantSlug := req.Header().Get("X-Tenant")
			authSub := req.Header().Get("X-User")

			scope, err := authUsecase.ResolveScope(ctx, tenantSlug, authSub)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			// Add scope to context
			newCtx := context.WithValue(ctx, scopeContextKey, *scope)
			return next(newCtx, req)
		}
	})
}

// GetScopeFromContext retrieves the domain.Scope from the context.
// It panics if the scope is not found, as it should always be present after the AuthInterceptor.
func GetScopeFromContext(ctx context.Context) domain.Scope {
	scope, ok := ctx.Value(scopeContextKey).(domain.Scope)
	if !ok {
		panic("scope not found in context")
	}
	return scope
}
