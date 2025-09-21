package port

import (
	"context"

	v1 "github.com/example/something-like-sns/apps/api/gen/sns/v1"
	"github.com/example/something-like-sns/apps/api/internal/domain"
)

// TimelineUsecase defines the input port for timeline-related operations.
type TimelineUsecase interface {
	CreatePost(ctx context.Context, scope domain.Scope, body string) (*domain.Post, error)
	ListFeed(ctx context.Context, scope domain.Scope, token string) ([]*domain.Post, string, error)
	CreateComment(ctx context.Context, scope domain.Scope, postID uint64, body string) (*domain.Comment, error)
    ListComments(ctx context.Context, scope domain.Scope, postID uint64, token string) ([]*domain.Comment, string, error)
}

// ReactionUsecase defines the input port for reaction-related operations.
type ReactionUsecase interface {
	ToggleReaction(ctx context.Context, scope domain.Scope, targetType v1.TargetType, targetID uint64, reactionType string) (*domain.Reaction, error)
}

// AuthUsecase defines the input port for authentication and authorization.
type AuthUsecase interface {
	ResolveScope(ctx context.Context, tenantSlug, userAuthSub string) (*domain.Scope, error)
	ResolveTenant(ctx context.Context, host string) (*domain.Tenant, error)
	GetMe(ctx context.Context, userID uint64) (*domain.User, error)
}

// DMUsecase defines the input port for DM-related operations.
type DMUsecase interface {
	GetOrCreateDM(ctx context.Context, scope domain.Scope, otherUserID uint64) (uint64, error)
	ListConversations(ctx context.Context, scope domain.Scope, token string) ([]*domain.Conversation, string, error)
	ListMessages(ctx context.Context, scope domain.Scope, conversationID uint64, token string) ([]*domain.Message, string, error)
	SendMessage(ctx context.Context, scope domain.Scope, conversationID uint64, body string) (*domain.Message, error)
}
