package service

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/example/something-like-sns/apps/api/gen/sns/v1"
	v1connect "github.com/example/something-like-sns/apps/api/gen/sns/v1/v1connect"
)

type TimelineServiceServer struct {
    db              *sql.DB
    allowDevHeaders bool
}

func NewTimelineServiceServer(db *sql.DB, allowDevHeaders bool) *TimelineServiceServer {
    return &TimelineServiceServer{db: db, allowDevHeaders: allowDevHeaders}
}

func (s *TimelineServiceServer) MountHandler() (string, http.Handler) {
    path, h := v1connect.NewTimelineServiceHandler(s)
    return path, h
}

func decodeCursor(token string) (time.Time, uint64, error) {
    if token == "" { return time.Time{}, 0, nil }
    raw, err := base64.StdEncoding.DecodeString(token)
    if err != nil { return time.Time{}, 0, err }
    parts := strings.SplitN(string(raw), ":", 2)
    if len(parts) != 2 { return time.Time{}, 0, errors.New("bad cursor") }
    t, err := time.Parse(time.RFC3339Nano, parts[0])
    if err != nil { return time.Time{}, 0, err }
    id, err := strconv.ParseUint(parts[1], 10, 64)
    return t, id, err
}

func encodeCursor(t time.Time, id uint64) *v1.Cursor {
    if t.IsZero() { return nil }
    token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%d", t.Format(time.RFC3339Nano), id)))
    return &v1.Cursor{Token: token}
}

// ListFeed: created_at DESC, id DESC
func (s *TimelineServiceServer) ListFeed(ctx context.Context, req *connect.Request[v1.ListFeedRequest]) (*connect.Response[v1.ListFeedResponse], error) {
    if !s.allowDevHeaders { return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("dev headers disabled")) }
    scope, err := resolveScopeFromDevHeaders(ctx, s.db, req.Header())
    if err != nil { return nil, err }

    var condCreated time.Time
    var condID uint64
    if c := req.Msg.GetCursor(); c != nil {
        condCreated, condID, err = decodeCursor(c.GetToken())
        if err != nil { return nil, connect.NewError(connect.CodeInvalidArgument, err) }
    }

    // Page size fixed minimal (20)
    const limit = 20
    var rows *sql.Rows
    if condID == 0 {
        rows, err = s.db.QueryContext(ctx, `
            SELECT p.id, p.author_user_id, p.body, p.created_at,
                   (SELECT COUNT(*) FROM reactions r WHERE r.tenant_id=p.tenant_id AND r.target_type='post' AND r.target_id=p.id) AS like_count,
                   (SELECT COUNT(*) FROM comments c WHERE c.tenant_id=p.tenant_id AND c.post_id=p.id) AS comment_count,
                   EXISTS(SELECT 1 FROM reactions r WHERE r.tenant_id=p.tenant_id AND r.target_type='post' AND r.target_id=p.id AND r.user_id=?) as liked
            FROM posts p
            WHERE p.tenant_id=? AND p.deleted_at IS NULL
            ORDER BY p.created_at DESC, p.id DESC
            LIMIT ?`, scope.UserID, scope.TenantID, limit)
    } else {
        rows, err = s.db.QueryContext(ctx, `
            SELECT p.id, p.author_user_id, p.body, p.created_at,
                   (SELECT COUNT(*) FROM reactions r WHERE r.tenant_id=p.tenant_id AND r.target_type='post' AND r.target_id=p.id) AS like_count,
                   (SELECT COUNT(*) FROM comments c WHERE c.tenant_id=p.tenant_id AND c.post_id=p.id) AS comment_count,
                   EXISTS(SELECT 1 FROM reactions r WHERE r.tenant_id=p.tenant_id AND r.target_type='post' AND r.target_id=p.id AND r.user_id=?) as liked
            FROM posts p
            WHERE p.tenant_id=? AND p.deleted_at IS NULL AND (p.created_at < ? OR (p.created_at = ? AND p.id < ?))
            ORDER BY p.created_at DESC, p.id DESC
            LIMIT ?`, scope.UserID, scope.TenantID, condCreated, condCreated, condID, limit)
    }
    if err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
    defer rows.Close()

    items := make([]*v1.Post, 0, limit)
    var lastT time.Time
    var lastID uint64
    for rows.Next() {
        var p v1.Post
        var created time.Time
        var liked bool
        var likeCount uint32
        var commentCount uint32
        if err := rows.Scan(&p.Id, &p.AuthorUserId, &p.Body, &created, &likeCount, &commentCount, &liked); err != nil {
            return nil, connect.NewError(connect.CodeInternal, err)
        }
        p.CreatedAt = created.Format(time.RFC3339Nano)
        p.LikedByMe = liked
        p.LikeCount = likeCount
        p.CommentCount = commentCount
        items = append(items, &p)
        lastT = created
        lastID = p.Id
    }
    if err := rows.Err(); err != nil { return nil, connect.NewError(connect.CodeInternal, err) }

    res := &v1.ListFeedResponse{Items: items}
    if len(items) == limit { res.Next = encodeCursor(lastT, lastID) }
    return connect.NewResponse(res), nil
}

func (s *TimelineServiceServer) CreatePost(ctx context.Context, req *connect.Request[v1.CreatePostRequest]) (*connect.Response[v1.CreatePostResponse], error) {
    if !s.allowDevHeaders { return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("dev headers disabled")) }
    scope, err := resolveScopeFromDevHeaders(ctx, s.db, req.Header())
    if err != nil { return nil, err }
    body := strings.TrimSpace(req.Msg.GetBody())
    if body == "" || len(body) > 2000 { return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid body")) }
    resExec, err := s.db.ExecContext(ctx, "INSERT INTO posts (tenant_id, author_user_id, body) VALUES (?,?,?)", scope.TenantID, scope.UserID, body)
    if err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
    id, _ := resExec.LastInsertId()
    var created time.Time
    _ = s.db.QueryRowContext(ctx, "SELECT created_at FROM posts WHERE id=?", id).Scan(&created)
    post := &v1.Post{Id: uint64(id), AuthorUserId: scope.UserID, Body: body, CreatedAt: created.Format(time.RFC3339Nano), LikedByMe: false, LikeCount: 0, CommentCount: 0}
    return connect.NewResponse(&v1.CreatePostResponse{Post: post}), nil
}

func (s *TimelineServiceServer) ListComments(ctx context.Context, req *connect.Request[v1.ListCommentsRequest]) (*connect.Response[v1.ListCommentsResponse], error) {
    if !s.allowDevHeaders { return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("dev headers disabled")) }
    scope, err := resolveScopeFromDevHeaders(ctx, s.db, req.Header())
    if err != nil { return nil, err }
    postID := req.Msg.GetPostId()
    const limit = 50
    rows, err := s.db.QueryContext(ctx, `
        SELECT id, author_user_id, body, created_at
        FROM comments
        WHERE tenant_id=? AND post_id=?
        ORDER BY created_at ASC, id ASC
        LIMIT ?`, scope.TenantID, postID, limit)
    if err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
    defer rows.Close()
    items := make([]*v1.Comment, 0, limit)
    for rows.Next() {
        var cmt v1.Comment
        var created time.Time
        if err := rows.Scan(&cmt.Id, &cmt.AuthorUserId, &cmt.Body, &created); err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
        cmt.PostId = postID
        cmt.CreatedAt = created.Format(time.RFC3339Nano)
        items = append(items, &cmt)
    }
    if err := rows.Err(); err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
    return connect.NewResponse(&v1.ListCommentsResponse{Items: items}), nil
}

func (s *TimelineServiceServer) CreateComment(ctx context.Context, req *connect.Request[v1.CreateCommentRequest]) (*connect.Response[v1.CreateCommentResponse], error) {
    if !s.allowDevHeaders { return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("dev headers disabled")) }
    scope, err := resolveScopeFromDevHeaders(ctx, s.db, req.Header())
    if err != nil { return nil, err }
    postID := req.Msg.GetPostId()
    body := strings.TrimSpace(req.Msg.GetBody())
    if body == "" || len(body) > 2000 { return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid body")) }
    resExec, err := s.db.ExecContext(ctx, "INSERT INTO comments (tenant_id, post_id, author_user_id, body) VALUES (?,?,?,?)", scope.TenantID, postID, scope.UserID, body)
    if err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
    id, _ := resExec.LastInsertId()
    var created time.Time
    _ = s.db.QueryRowContext(ctx, "SELECT created_at FROM comments WHERE id=?", id).Scan(&created)
    cmt := &v1.Comment{Id: uint64(id), PostId: postID, AuthorUserId: scope.UserID, Body: body, CreatedAt: created.Format(time.RFC3339Nano)}
    return connect.NewResponse(&v1.CreateCommentResponse{Comment: cmt}), nil
}


