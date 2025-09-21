package rpc

import (
    "context"
    "net/http"

    "connectrpc.com/connect"
    v1 "github.com/example/something-like-sns/apps/api/gen/sns/v1"
    "github.com/example/something-like-sns/apps/api/gen/sns/v1/v1connect"
    "github.com/example/something-like-sns/apps/api/internal/port"
)

type ReactionHandler struct {
	reactionUsecase port.ReactionUsecase
}

func NewReactionHandler(ru port.ReactionUsecase) *ReactionHandler {
	return &ReactionHandler{reactionUsecase: ru}
}

func (s *ReactionHandler) MountHandler(authInterceptor connect.Interceptor) (string, http.Handler) {
	path, h := v1connect.NewReactionServiceHandler(s, connect.WithInterceptors(authInterceptor))
	return path, h
}

func (s *ReactionHandler) ToggleReaction(ctx context.Context, req *connect.Request[v1.ToggleReactionRequest]) (*connect.Response[v1.ToggleReactionResponse], error) {
	scope := GetScopeFromContext(ctx)

	reaction, err := s.reactionUsecase.ToggleReaction(ctx, scope, req.Msg.GetTargetType(), req.Msg.GetTargetId(), req.Msg.GetType())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.ToggleReactionResponse{
		Active: reaction.Active,
		Total:  reaction.Total,
	}), nil
}
