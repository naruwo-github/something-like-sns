package port

import (
	"context"
	"time"

	"github.com/example/something-like-sns/apps/api/internal/domain"
)

// TimelineRepository defines the output port for timeline data persistence.
type TimelineRepository interface {
	CreatePost(ctx context.Context, tenantID, authorID uint64, body string) (*domain.Post, error)
	FindFeed(ctx context.Context, tenantID, userID uint64, limit int, cursorTime time.Time, cursorID uint64) ([]*domain.Post, error)
	CreateComment(ctx context.Context, tenantID, postID, authorID uint64, body string) (*domain.Comment, error)
    FindCommentsByPostID(ctx context.Context, tenantID, postID uint64, limit int, cursorTime time.Time, cursorID uint64) ([]*domain.Comment, error)
}

// ReactionRepository defines the output port for reaction data persistence.
type ReactionRepository interface {
	Toggle(ctx context.Context, tenantID, userID uint64, targetType domain.ReactionTargetType, targetID uint64, reactionType string) (bool, error)
	Count(ctx context.Context, tenantID uint64, targetType domain.ReactionTargetType, targetID uint64, reactionType string) (uint32, error)
}

// AuthRepository defines the output port for user and tenant data persistence.
type AuthRepository interface {
	FindTenantByHost(ctx context.Context, host string) (*domain.Tenant, error)
	FindTenantBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
	FindOrCreateUser(ctx context.Context, authSub, displayName string) (uint64, error)
	FindUserByID(ctx context.Context, userID uint64) (*domain.User, error)
	EnsureMembership(ctx context.Context, tenantID, userID uint64, role string) error
	FindUserMemberships(ctx context.Context, userID uint64) ([]*domain.TenantMembership, error)
}

// DMRepository defines the output port for DM data persistence.
type DMRepository interface {
	FindDMConversation(ctx context.Context, tenantID, userID1, userID2 uint64) (uint64, error)
	CreateDMConversation(ctx context.Context, tenantID uint64, userIDs ...uint64) (uint64, error)
	FindConversations(ctx context.Context, tenantID, userID uint64, limit int, cursorTime time.Time, cursorID uint64) ([]*domain.Conversation, error)
	FindMessages(ctx context.Context, tenantID, conversationID uint64, limit int, cursorTime time.Time, cursorID uint64) ([]*domain.Message, error)
	CreateMessage(ctx context.Context, tenantID, conversationID, senderID uint64, body string) (*domain.Message, error)
}

// Store defines the interface for accessing all repositories.
// It also provides a method to execute operations within a database transaction.
type Store interface {
	AuthRepository() AuthRepository
	TimelineRepository() TimelineRepository
	ReactionRepository() ReactionRepository
	DMRepository() DMRepository
	ExecTx(ctx context.Context, fn func(Store) error) error
}
