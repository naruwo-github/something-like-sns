package service

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	v1 "github.com/example/something-like-sns/apps/api/gen/sns/v1"
	v1connect "github.com/example/something-like-sns/apps/api/gen/sns/v1/v1connect"
)

type ReactionServiceServer struct {
    db              *sql.DB
    allowDevHeaders bool
}

func NewReactionServiceServer(db *sql.DB, allowDevHeaders bool) *ReactionServiceServer {
    return &ReactionServiceServer{db: db, allowDevHeaders: allowDevHeaders}
}

func (s *ReactionServiceServer) MountHandler() (string, http.Handler) {
    path, h := v1connect.NewReactionServiceHandler(s)
    return path, h
}

func (s *ReactionServiceServer) ToggleReaction(ctx context.Context, req *connect.Request[v1.ToggleReactionRequest]) (*connect.Response[v1.ToggleReactionResponse], error) {
    if !s.allowDevHeaders { return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("dev headers disabled")) }
    scope, err := resolveScopeFromDevHeaders(ctx, s.db, req.Header())
    if err != nil { return nil, err }
    tt := req.Msg.GetTargetType()
    id := req.Msg.GetTargetId()
    typ := req.Msg.GetType()
    if typ == "" { typ = "like" }
    var target string
    switch tt {
    case v1.TargetType_POST:
        target = "post"
    case v1.TargetType_COMMENT:
        target = "comment"
    default:
        return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid target type"))
    }

    // Try delete; if zero rows affected, insert
    res, err := s.db.ExecContext(ctx, "DELETE FROM reactions WHERE tenant_id=? AND target_type=? AND target_id=? AND user_id=? AND type=?", scope.TenantID, target, id, scope.UserID, typ)
    if err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
    affected, _ := res.RowsAffected()
    active := false
    if affected == 0 {
        if _, err := s.db.ExecContext(ctx, "INSERT INTO reactions (tenant_id, target_type, target_id, user_id, type) VALUES (?,?,?,?,?)", scope.TenantID, target, id, scope.UserID, typ); err != nil {
            return nil, connect.NewError(connect.CodeInternal, err)
        }
        active = true
    }
    var total uint32
    _ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM reactions WHERE tenant_id=? AND target_type=? AND target_id=? AND type=?", scope.TenantID, target, id, typ).Scan(&total)
    return connect.NewResponse(&v1.ToggleReactionResponse{Active: active, Total: total}), nil
}


