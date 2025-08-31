package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/example/something-like-sns/apps/api/internal/domain"
)

type dmRepository struct {
	q DBTX
}

func (r *dmRepository) FindDMConversation(ctx context.Context, tenantID, userID1, userID2 uint64) (uint64, error) {
	var convID uint64
	q := `SELECT c.id FROM conversations c
          JOIN conversation_members m1 ON m1.conversation_id=c.id AND m1.user_id=?
          JOIN conversation_members m2 ON m2.conversation_id=c.id AND m2.user_id=?
          WHERE c.tenant_id=? AND c.kind='dm' LIMIT 1`
	err := r.q.QueryRowContext(ctx, q, userID1, userID2, tenantID).Scan(&convID)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return convID, nil
}

func (r *dmRepository) CreateDMConversation(ctx context.Context, tenantID uint64, userIDs ...uint64) (uint64, error) {
	// This method will be called within a transaction from the usecase layer.
	// The transaction is handled by the sqlStore.
	res, err := r.q.ExecContext(ctx, "INSERT INTO conversations (tenant_id, kind) VALUES (?, 'dm')", tenantID)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	convID := uint64(id)

	valueStrings := make([]string, 0, len(userIDs))
	valueArgs := make([]interface{}, 0, len(userIDs)*2)
	for _, userID := range userIDs {
		valueStrings = append(valueStrings, "(?, ?)")
		valueArgs = append(valueArgs, convID, userID)
	}
	stmt := fmt.Sprintf("INSERT INTO conversation_members (conversation_id, user_id) VALUES %s", strings.Join(valueStrings, ","))
	
	if _, err := r.q.ExecContext(ctx, stmt, valueArgs...); err != nil {
		return 0, err
	}

	return convID, nil
}

func (r *dmRepository) FindConversations(ctx context.Context, tenantID, userID uint64, limit int, cursorTime time.Time, cursorID uint64) ([]*domain.Conversation, error) {
	// For simplicity, cursor is ignored in this implementation for conversations
	rows, err := r.q.QueryContext(ctx, `
        SELECT c.id, c.created_at
        FROM conversations c
        JOIN conversation_members m ON m.conversation_id=c.id AND m.user_id=?
        WHERE c.tenant_id=?
        ORDER BY c.created_at DESC, c.id DESC
        LIMIT ?`, userID, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*domain.Conversation, 0, limit)
	for rows.Next() {
		var conv domain.Conversation
		if err := rows.Scan(&conv.ID, &conv.CreatedAt); err != nil {
			return nil, err
		}
		// Fetch members
		mrows, err := r.q.QueryContext(ctx, "SELECT user_id FROM conversation_members WHERE conversation_id=? ORDER BY user_id", conv.ID)
		if err != nil {
			return nil, err
		}
		var members []uint64
		for mrows.Next() {
			var uid uint64
			if err := mrows.Scan(&uid); err != nil {
				_ = mrows.Close()
				return nil, err
			}
			members = append(members, uid)
		}
		mrows.Close()
		conv.MemberUserIDs = members
		items = append(items, &conv)
	}
	return items, rows.Err()
}

func (r *dmRepository) FindMessages(ctx context.Context, tenantID, conversationID uint64, limit int, cursorTime time.Time, cursorID uint64) ([]*domain.Message, error) {
	var rows *sql.Rows
	var err error
	if cursorID == 0 {
		rows, err = r.q.QueryContext(ctx, `
            SELECT id, sender_user_id, body, created_at
            FROM messages
            WHERE tenant_id=? AND conversation_id=?
            ORDER BY created_at DESC, id DESC
            LIMIT ?`, tenantID, conversationID, limit)
	} else {
		rows, err = r.q.QueryContext(ctx, `
            SELECT id, sender_user_id, body, created_at
            FROM messages
            WHERE tenant_id=? AND conversation_id=? AND (created_at < ? OR (created_at = ? AND id < ?))
            ORDER BY created_at DESC, id DESC
            LIMIT ?`, tenantID, conversationID, cursorTime, cursorTime, cursorID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*domain.Message, 0, limit)
	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(&msg.ID, &msg.SenderUserID, &msg.Body, &msg.CreatedAt); err != nil {
			return nil, err
		}
		msg.ConversationID = conversationID
		items = append(items, &msg)
	}
	return items, rows.Err()
}

func (r *dmRepository) CreateMessage(ctx context.Context, tenantID, conversationID, senderID uint64, body string) (*domain.Message, error) {
	resExec, err := r.q.ExecContext(ctx, "INSERT INTO messages (tenant_id, conversation_id, sender_user_id, body) VALUES (?,?,?,?)", tenantID, conversationID, senderID, body)
	if err != nil {
		return nil, err
	}
	id, _ := resExec.LastInsertId()
	var created time.Time
	_ = r.q.QueryRowContext(ctx, "SELECT created_at FROM messages WHERE id=?", id).Scan(&created)
	return &domain.Message{
		ID:             uint64(id),
		ConversationID: conversationID,
		SenderUserID:   senderID,
		Body:           body,
		CreatedAt:      created,
	}, nil
}
