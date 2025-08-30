package port

import (
	"context"
	"time"

	"github.com/example/something-like-sns/apps/api/gen/sns/v1"
	"github.com/example/something-like-sns/apps/api/internal/domain"
)

// TimelineUsecase defines the input port for timeline-related operations.
type TimelineUsecase interface {
	CreatePost(ctx context.Context, scope domain.Scope, body string) (*domain.Post, error)
	ListFeed(ctx context.Context, scope domain.Scope, token string) ([]*domain.Post, string, error)
	CreateComment(ctx context.Context, scope domain.Scope, postID uint64, body string) (*domain.Comment, error)
	ListComments(ctx context.Context, scope domain.Scope, postID uint64) ([]*domain.Comment, error)
}

// TimelineRepository defines the output port for timeline data persistence.
type TimelineRepository interface {
	CreatePost(ctx context.Context, tenantID, authorID uint64, body string) (*domain.Post, error)
	FindFeed(ctx context.Context, tenantID, userID uint64, limit int, cursorTime time.Time, cursorID uint64) ([]*domain.Post, error)
	CreateComment(ctx context.Context, tenantID, postID, authorID uint64, body string) (*domain.Comment, error)
	FindCommentsByPostID(ctx context.Context, tenantID, postID uint64, limit int) ([]*domain.Comment, error)
}

// ReactionUsecase defines the input port for reaction-related operations.
type ReactionUsecase interface {
	ToggleReaction(ctx context.Context, scope domain.Scope, targetType v1.TargetType, targetID uint64, reactionType string) (*domain.Reaction, error)
}

// ReactionRepository defines the output port for reaction data persistence.
type ReactionRepository interface {
	Toggle(ctx context.Context, tenantID, userID uint64, targetType domain.ReactionTargetType, targetID uint64, reactionType string) (bool, error)
	Count(ctx context.Context, tenantID uint64, targetType domain.ReactionTargetType, targetID uint64, reactionType string) (uint32, error)
}

// AuthUsecase defines the input port for authentication and authorization.
type AuthUsecase interface {
	ResolveScope(ctx context.Context, tenantSlug, userAuthSub string) (*domain.Scope, error)
	ResolveTenant(ctx context.Context, host string) (*domain.Tenant, error)
	GetMe(ctx context.Context, tenantSlug, userAuthSub string) (*domain.User, error)
}

// AuthRepository defines the output port for user and tenant data persistence.
type AuthRepository interface {
	FindTenantByHost(ctx context.Context, host string) (*domain.Tenant, error)
	FindTenantBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
	FindOrCreateUser(ctx context.Context, authSub, displayName string) (uint64, error)
	EnsureMembership(ctx context.Context, tenantID, userID uint64, role string) error
	FindUserMemberships(ctx context.Context, userID uint64) ([]*domain.TenantMembership, error)
}


// CursorEncoder defines an interface for encoding and decoding cursors.
type CursorEncoder interface {
	Encode(t time.Time, id uint64) string
	Decode(token string) (time.Time, uint64, error)
}