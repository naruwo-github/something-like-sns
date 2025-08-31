package application

import (
	"context"
	"errors"

	v1 "github.com/example/something-like-sns/apps/api/gen/sns/v1"
	"github.com/example/something-like-sns/apps/api/internal/domain"
	"github.com/example/something-like-sns/apps/api/internal/port"
)

type reactionUsecase struct {
	store port.Store
}

func NewReactionUsecase(store port.Store) port.ReactionUsecase {
	return &reactionUsecase{store: store}
}

func (u *reactionUsecase) ToggleReaction(ctx context.Context, scope domain.Scope, targetType v1.TargetType, targetID uint64, reactionType string) (*domain.Reaction, error) {
	if reactionType == "" {
		reactionType = "like"
	}

	var domainTargetType domain.ReactionTargetType
	switch targetType {
	case v1.TargetType_POST:
		domainTargetType = domain.ReactionTargetPost
	case v1.TargetType_COMMENT:
		domainTargetType = domain.ReactionTargetComment
	default:
		return nil, errors.New("invalid target type")
	}

	active, err := u.store.ReactionRepository().Toggle(ctx, scope.TenantID, scope.UserID, domainTargetType, targetID, reactionType)
	if err != nil {
		return nil, err
	}

	total, err := u.store.ReactionRepository().Count(ctx, scope.TenantID, domainTargetType, targetID, reactionType)
	if err != nil {
		return nil, err
	}

	return &domain.Reaction{Active: active, Total: total}, nil
}
