package rpc

import (
	"context"
	"errors"
	"net/http"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/example/something-like-sns/apps/api/gen/sns/v1"
	"github.com/example/something-like-sns/apps/api/gen/sns/v1/v1connect"
	"github.com/example/something-like-sns/apps/api/internal/port"
)

type TimelineHandler struct {
	authUsecase     port.AuthUsecase
	timelineUsecase port.TimelineUsecase
	allowDevHeaders bool
}

func NewTimelineHandler(au port.AuthUsecase, tu port.TimelineUsecase, allowDev bool) *TimelineHandler {
	return &TimelineHandler{authUsecase: au, timelineUsecase: tu, allowDevHeaders: allowDev}
}

func (s *TimelineHandler) MountHandler() (string, http.Handler) {
	path, h := v1connect.NewTimelineServiceHandler(s)
	return path, h
}

func (s *TimelineHandler) getScope(ctx context.Context, h http.Header) (domain.Scope, error) {
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

func (s *TimelineHandler) ListFeed(ctx context.Context, req *connect.Request[v1.ListFeedRequest]) (*connect.Response[v1.ListFeedResponse], error) {
	scope, err := s.getScope(ctx, req.Header())
	if err != nil {
		return nil, err
	}

	posts, nextToken, err := s.timelineUsecase.ListFeed(ctx, scope, req.Msg.GetCursor().GetToken())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	items := make([]*v1.Post, len(posts))
	for i, p := range posts {
		items[i] = &v1.Post{
			Id:           p.ID,
			AuthorUserId: p.AuthorUserID,
			Body:         p.Body,
			CreatedAt:    p.CreatedAt.Format(time.RFC3339Nano),
			LikedByMe:    p.LikedByMe,
			LikeCount:    p.LikeCount,
			CommentCount: p.CommentCount,
		}
	}

	res := &v1.ListFeedResponse{Items: items}
	if nextToken != "" {
		res.Next = &v1.Cursor{Token: nextToken}
	}
	return connect.NewResponse(res), nil
}

func (s *TimelineHandler) CreatePost(ctx context.Context, req *connect.Request[v1.CreatePostRequest]) (*connect.Response[v1.CreatePostResponse], error) {
	scope, err := s.getScope(ctx, req.Header())
	if err != nil {
		return nil, err
	}

	post, err := s.timelineUsecase.CreatePost(ctx, scope, req.Msg.GetBody())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&v1.CreatePostResponse{
		Post: &v1.Post{
			Id:           post.ID,
			AuthorUserId: post.AuthorUserID,
			Body:         post.Body,
			CreatedAt:    post.CreatedAt.Format(time.RFC3339Nano),
		},
	}), nil
}

func (s *TimelineHandler) ListComments(ctx context.Context, req *connect.Request[v1.ListCommentsRequest]) (*connect.Response[v1.ListCommentsResponse], error) {
	scope, err := s.getScope(ctx, req.Header())
	if err != nil {
		return nil, err
	}

	comments, err := s.timelineUsecase.ListComments(ctx, scope, req.Msg.GetPostId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	items := make([]*v1.Comment, len(comments))
	for i, c := range comments {
		items[i] = &v1.Comment{
			Id:           c.ID,
			PostId:       c.PostID,
			AuthorUserId: c.AuthorUserID,
			Body:         c.Body,
			CreatedAt:    c.CreatedAt.Format(time.RFC3339Nano),
		}
	}

	return connect.NewResponse(&v1.ListCommentsResponse{Items: items}), nil
}

func (s *TimelineHandler) CreateComment(ctx context.Context, req *connect.Request[v1.CreateCommentRequest]) (*connect.Response[v1.CreateCommentResponse], error) {
	scope, err := s.getScope(ctx, req.Header())
	if err != nil {
		return nil, err
	}

	comment, err := s.timelineUsecase.CreateComment(ctx, scope, req.Msg.GetPostId(), req.Msg.GetBody())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&v1.CreateCommentResponse{
		Comment: &v1.Comment{
			Id:           comment.ID,
			PostId:       comment.PostID,
			AuthorUserId: comment.AuthorUserID,
			Body:         comment.Body,
			CreatedAt:    comment.CreatedAt.Format(time.RFC3339Nano),
		},
	}), nil
}
