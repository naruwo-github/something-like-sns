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

type DMServiceServer struct {
    db              *sql.DB
    allowDevHeaders bool
}

func NewDMServiceServer(db *sql.DB, allowDevHeaders bool) *DMServiceServer {
    return &DMServiceServer{db: db, allowDevHeaders: allowDevHeaders}
}

func (s *DMServiceServer) MountHandler() (string, http.Handler) {
    path, h := v1connect.NewDMServiceHandler(s)
    return path, h
}

func decodeCursorDM(token string) (time.Time, uint64, error) {
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

func encodeCursorDM(t time.Time, id uint64) *v1.Cursor {
    if t.IsZero() { return nil }
    token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%d", t.Format(time.RFC3339Nano), id)))
    return &v1.Cursor{Token: token}
}

func (s *DMServiceServer) GetOrCreateDM(ctx context.Context, req *connect.Request[v1.GetOrCreateDMRequest]) (*connect.Response[v1.GetOrCreateDMResponse], error) {
    if !s.allowDevHeaders { return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("dev headers disabled")) }
    scope, err := resolveScopeFromDevHeaders(ctx, s.db, req.Header())
    if err != nil { return nil, err }
    other := req.Msg.GetOtherUserId()
    if other == 0 || other == scope.UserID { return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid other_user_id")) }

    // Find existing conversation for both members
    var convID uint64
    q := `SELECT c.id FROM conversations c
          JOIN conversation_members m1 ON m1.conversation_id=c.id AND m1.user_id=?
          JOIN conversation_members m2 ON m2.conversation_id=c.id AND m2.user_id=?
          WHERE c.tenant_id=? AND c.kind='dm' LIMIT 1`
    err = s.db.QueryRowContext(ctx, q, scope.UserID, other, scope.TenantID).Scan(&convID)
    if err != nil && err != sql.ErrNoRows { return nil, connect.NewError(connect.CodeInternal, err) }
    if convID == 0 {
        tx, err := s.db.BeginTx(ctx, nil)
        if err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
        res, err := tx.ExecContext(ctx, "INSERT INTO conversations (tenant_id, kind) VALUES (?, 'dm')", scope.TenantID)
        if err != nil { _ = tx.Rollback(); return nil, connect.NewError(connect.CodeInternal, err) }
        id, _ := res.LastInsertId()
        convID = uint64(id)
        if _, err := tx.ExecContext(ctx, "INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?), (?, ?)", convID, scope.UserID, convID, other); err != nil {
            _ = tx.Rollback(); return nil, connect.NewError(connect.CodeInternal, err)
        }
        if err := tx.Commit(); err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
    }
    return connect.NewResponse(&v1.GetOrCreateDMResponse{ConversationId: convID}), nil
}

func (s *DMServiceServer) ListConversations(ctx context.Context, req *connect.Request[v1.ListConversationsRequest]) (*connect.Response[v1.ListConversationsResponse], error) {
    if !s.allowDevHeaders { return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("dev headers disabled")) }
    scope, err := resolveScopeFromDevHeaders(ctx, s.db, req.Header())
    if err != nil { return nil, err }

    const limit = 20
    rows, err := s.db.QueryContext(ctx, `
        SELECT c.id, c.created_at
        FROM conversations c
        JOIN conversation_members m ON m.conversation_id=c.id AND m.user_id=?
        WHERE c.tenant_id=?
        ORDER BY c.created_at DESC, c.id DESC
        LIMIT ?`, scope.UserID, scope.TenantID, limit)
    if err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
    defer rows.Close()

    items := make([]*v1.Conversation, 0, limit)
    var lastT time.Time
    var lastID uint64
    for rows.Next() {
        var id uint64
        var created time.Time
        if err := rows.Scan(&id, &created); err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
        // Fetch members
        mrows, err := s.db.QueryContext(ctx, "SELECT user_id FROM conversation_members WHERE conversation_id=? ORDER BY user_id", id)
        if err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
        var members []uint64
        for mrows.Next() {
            var uid uint64
            if err := mrows.Scan(&uid); err != nil { _ = mrows.Close(); return nil, connect.NewError(connect.CodeInternal, err) }
            members = append(members, uid)
        }
        mrows.Close()
        items = append(items, &v1.Conversation{Id: id, CreatedAt: created.Format(time.RFC3339Nano), MemberUserIds: members})
        lastT = created
        lastID = id
    }
    if err := rows.Err(); err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
    res := &v1.ListConversationsResponse{Items: items}
    if len(items) == limit { res.Next = encodeCursorDM(lastT, lastID) }
    return connect.NewResponse(res), nil
}

func (s *DMServiceServer) ListMessages(ctx context.Context, req *connect.Request[v1.ListMessagesRequest]) (*connect.Response[v1.ListMessagesResponse], error) {
    if !s.allowDevHeaders { return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("dev headers disabled")) }
    scope, err := resolveScopeFromDevHeaders(ctx, s.db, req.Header())
    if err != nil { return nil, err }
    cnv := req.Msg.GetConversationId()

    var condCreated time.Time
    var condID uint64
    if c := req.Msg.GetCursor(); c != nil {
        condCreated, condID, err = decodeCursorDM(c.GetToken())
        if err != nil { return nil, connect.NewError(connect.CodeInvalidArgument, err) }
    }
    const limit = 50
    var rows *sql.Rows
    if condID == 0 {
        rows, err = s.db.QueryContext(ctx, `
            SELECT id, sender_user_id, body, created_at
            FROM messages
            WHERE tenant_id=? AND conversation_id=?
            ORDER BY created_at DESC, id DESC
            LIMIT ?`, scope.TenantID, cnv, limit)
    } else {
        rows, err = s.db.QueryContext(ctx, `
            SELECT id, sender_user_id, body, created_at
            FROM messages
            WHERE tenant_id=? AND conversation_id=? AND (created_at < ? OR (created_at = ? AND id < ?))
            ORDER BY created_at DESC, id DESC
            LIMIT ?`, scope.TenantID, cnv, condCreated, condCreated, condID, limit)
    }
    if err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
    defer rows.Close()
    items := make([]*v1.Message, 0, limit)
    var lastT time.Time
    var lastID uint64
    for rows.Next() {
        var id uint64
        var sender uint64
        var body string
        var created time.Time
        if err := rows.Scan(&id, &sender, &body, &created); err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
        items = append(items, &v1.Message{Id: id, ConversationId: cnv, SenderUserId: sender, Body: body, CreatedAt: created.Format(time.RFC3339Nano)})
        lastT = created
        lastID = id
    }
    if err := rows.Err(); err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
    res := &v1.ListMessagesResponse{Items: items}
    if len(items) == limit { res.Next = encodeCursorDM(lastT, lastID) }
    return connect.NewResponse(res), nil
}

func (s *DMServiceServer) SendMessage(ctx context.Context, req *connect.Request[v1.SendMessageRequest]) (*connect.Response[v1.SendMessageResponse], error) {
    if !s.allowDevHeaders { return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("dev headers disabled")) }
    scope, err := resolveScopeFromDevHeaders(ctx, s.db, req.Header())
    if err != nil { return nil, err }
    cnv := req.Msg.GetConversationId()
    body := strings.TrimSpace(req.Msg.GetBody())
    if body == "" || len(body) > 2000 { return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid body")) }
    resExec, err := s.db.ExecContext(ctx, "INSERT INTO messages (tenant_id, conversation_id, sender_user_id, body) VALUES (?,?,?,?)", scope.TenantID, cnv, scope.UserID, body)
    if err != nil { return nil, connect.NewError(connect.CodeInternal, err) }
    id, _ := resExec.LastInsertId()
    var created time.Time
    _ = s.db.QueryRowContext(ctx, "SELECT created_at FROM messages WHERE id=?", id).Scan(&created)
    msg := &v1.Message{Id: uint64(id), ConversationId: cnv, SenderUserId: scope.UserID, Body: body, CreatedAt: created.Format(time.RFC3339Nano)}
    return connect.NewResponse(&v1.SendMessageResponse{Message: msg}), nil
}


