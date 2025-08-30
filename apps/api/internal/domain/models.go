package domain

import "time"

// Scope represents the context of a request, typically derived from auth headers.
type Scope struct {
	TenantID uint64
	UserID   uint64
}

// Post represents a post in the system.
type Post struct {
	ID           uint64
	AuthorUserID uint64
	Body         string
	CreatedAt    time.Time
	LikedByMe    bool
	LikeCount    uint32
	CommentCount uint32
}

// Comment represents a comment on a post.
type Comment struct {
	ID           uint64
	PostID       uint64
	AuthorUserID uint64
	Body         string
	CreatedAt    time.Time
}

// ReactionTargetType defines the type of entity a reaction can be attached to.
type ReactionTargetType string

const (
	ReactionTargetPost    ReactionTargetType = "post"
	ReactionTargetComment ReactionTargetType = "comment"
)

// Reaction represents a reaction from a user to a target entity.
type Reaction struct {
	Active bool
	Total  uint32
}

// User represents a user in the system.
type User struct {
	ID          uint64
	DisplayName string
	Memberships []*TenantMembership
}

// TenantMembership represents a user's membership in a tenant.
type TenantMembership struct {
	TenantID   uint64
	TenantSlug string
	Role       string
}

// Tenant represents a tenant in the system.
type Tenant struct {
	ID   uint64
	Slug string
}

// Conversation represents a DM conversation.
type Conversation struct {
	ID            uint64
	CreatedAt     time.Time
	MemberUserIDs []uint64
}

// Message represents a message in a conversation.
type Message struct {
	ID             uint64
	ConversationID uint64
	SenderUserID   uint64
	Body           string
	CreatedAt      time.Time
}
