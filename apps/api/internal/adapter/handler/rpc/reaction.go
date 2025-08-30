package rpc

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	v1 "github.com/example/something-like-sns/apps/api/gen/sns/v1"
	"github.com/example/something-like-sns/apps/api/gen/sns/v1/v1connect"
	"github.com/example/something-like-sns/apps/api/internal/domain"
	"github.com/example/something-like-sns/apps/api/internal/port"
)

type ReactionHandler struct {
	authUsecase     port.AuthUsecase
	reactionUsecase port.ReactionUsecase
	allowDevHeaders bool
}

func NewReactionHandler(au port.AuthUsecase, ru port.ReactionUsecase, allowDev bool) *ReactionHandler {
	return &ReactionHandler{authUsecase: au, reactionUsecase: ru, allowDevHeaders: allowDev}
}

func (s *ReactionHandler) MountHandler() (string, http.Handler) {
	path, h := v1connect.NewReactionServiceHandler(s)
	return path, h
}

func (s *ReactionHandler) getScope(ctx context.Context, h http.Header) (domain.Scope, error) {
	if !s.allowDevHeaders {
		return domain.Scope{}, connect.NewError(connect.CodeUnauthenticated, errors.New("dev headers disabled"))
	}
	tenantSlug := h.Get("X-Tenant")
	authSub := h.Get("X-User")
	scope, err := s.authUsecase.ResolveScope(ctx, tenantSlug, authSub)
	if err != nil {
		return domain.Scope{}, connect.NewError(connect.CodeUnauthenticated, err)
	}
	return *scope, nil
}

func (s *ReactionHandler) ToggleReaction(ctx context.Context, req *connect.Request[v1.ToggleReactionRequest]) (*connect.Response[v1.ToggleReactionResponse], error) {
	scope, err := s.getScope(ctx, req.Header())
	if err != nil {
		return nil, err
	}

	reaction, err := s.reactionUsecase.ToggleReaction(ctx, scope, req.Msg.GetTargetType(), req.Msg.GetTargetId(), req.Msg.GetType())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.ToggleReactionResponse{
		Active: reaction.Active,
		Total:  reaction.Total,
	}), nil
}
