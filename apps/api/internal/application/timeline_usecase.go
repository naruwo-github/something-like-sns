package application

import (
	"context"
	"errors"
	"strings"

	"github.com/example/something-like-sns/apps/api/internal/domain"
	"github.com/example/something-like-sns/apps/api/internal/port"
)

type timelineUsecase struct {
	store         port.Store
	cursorEncoder port.CursorEncoder
}

func NewTimelineUsecase(store port.Store, ce port.CursorEncoder) port.TimelineUsecase {
	return &timelineUsecase{store: store, cursorEncoder: ce}
}

func (u *timelineUsecase) CreatePost(ctx context.Context, scope domain.Scope, body string) (*domain.Post, error) {
	body = strings.TrimSpace(body)
	if body == "" || len(body) > 2000 {
		return nil, errors.New("invalid body")
	}
	return u.store.TimelineRepository().CreatePost(ctx, scope.TenantID, scope.UserID, body)
}

func (u *timelineUsecase) ListFeed(ctx context.Context, scope domain.Scope, token string) ([]*domain.Post, string, error) {
	const limit = 20
	cursorTime, cursorID, err := u.cursorEncoder.Decode(token)
	if err != nil {
		return nil, "", err
	}

	posts, err := u.store.TimelineRepository().FindFeed(ctx, scope.TenantID, scope.UserID, limit, cursorTime, cursorID)
	if err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(posts) == limit {
		lastPost := posts[len(posts)-1]
		nextToken = u.cursorEncoder.Encode(lastPost.CreatedAt, lastPost.ID)
	}

	return posts, nextToken, nil
}

func (u *timelineUsecase) CreateComment(ctx context.Context, scope domain.Scope, postID uint64, body string) (*domain.Comment, error) {
	body = strings.TrimSpace(body)
	if body == "" || len(body) > 2000 {
		return nil, errors.New("invalid body")
	}
	return u.store.TimelineRepository().CreateComment(ctx, scope.TenantID, postID, scope.UserID, body)
}

func (u *timelineUsecase) ListComments(ctx context.Context, scope domain.Scope, postID uint64, token string) ([]*domain.Comment, string, error) {
    const limit = 50
    cursorTime, cursorID, err := u.cursorEncoder.Decode(token)
    if err != nil {
        return nil, "", err
    }
    comments, err := u.store.TimelineRepository().FindCommentsByPostID(ctx, scope.TenantID, postID, limit, cursorTime, cursorID)
    if err != nil {
        return nil, "", err
    }
    var nextToken string
    if len(comments) == limit {
        last := comments[len(comments)-1]
        nextToken = u.cursorEncoder.Encode(last.CreatedAt, last.ID)
    }
    return comments, nextToken, nil
}
