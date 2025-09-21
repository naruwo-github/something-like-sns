package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/example/something-like-sns/apps/api/internal/domain"
)

type timelineRepository struct {
	q DBTX
}

func (r *timelineRepository) CreatePost(ctx context.Context, tenantID, authorID uint64, body string) (*domain.Post, error) {
	resExec, err := r.q.ExecContext(ctx, "INSERT INTO posts (tenant_id, author_user_id, body) VALUES (?,?,?)", tenantID, authorID, body)
	if err != nil {
		return nil, err
	}
	id, _ := resExec.LastInsertId()
	var created time.Time
	_ = r.q.QueryRowContext(ctx, "SELECT created_at FROM posts WHERE id=?", id).Scan(&created)
	return &domain.Post{
		ID:           uint64(id),
		AuthorUserID: authorID,
		Body:         body,
		CreatedAt:    created,
	}, nil
}

func (r *timelineRepository) FindFeed(ctx context.Context, tenantID, userID uint64, limit int, cursorTime time.Time, cursorID uint64) ([]*domain.Post, error) {
	var rows *sql.Rows
	var err error
	if cursorID == 0 {
		rows, err = r.q.QueryContext(ctx, `
            SELECT p.id, p.author_user_id, p.body, p.created_at,
                   (SELECT COUNT(*) FROM reactions r WHERE r.tenant_id=p.tenant_id AND r.target_type='post' AND r.target_id=p.id) AS like_count,
                   (SELECT COUNT(*) FROM comments c WHERE c.tenant_id=p.tenant_id AND c.post_id=p.id) AS comment_count,
                   EXISTS(SELECT 1 FROM reactions r WHERE r.tenant_id=p.tenant_id AND r.target_type='post' AND r.target_id=p.id AND r.user_id=?) as liked
            FROM posts p
            WHERE p.tenant_id=? AND p.deleted_at IS NULL
            ORDER BY p.created_at DESC, p.id DESC
            LIMIT ?`, userID, tenantID, limit)
	} else {
		rows, err = r.q.QueryContext(ctx, `
            SELECT p.id, p.author_user_id, p.body, p.created_at,
                   (SELECT COUNT(*) FROM reactions r WHERE r.tenant_id=p.tenant_id AND r.target_type='post' AND r.target_id=p.id) AS like_count,
                   (SELECT COUNT(*) FROM comments c WHERE c.tenant_id=p.tenant_id AND c.post_id=p.id) AS comment_count,
                   EXISTS(SELECT 1 FROM reactions r WHERE r.tenant_id=p.tenant_id AND r.target_type='post' AND r.target_id=p.id AND r.user_id=?) as liked
            FROM posts p
            WHERE p.tenant_id=? AND p.deleted_at IS NULL AND (p.created_at < ? OR (p.created_at = ? AND p.id < ?))
            ORDER BY p.created_at DESC, p.id DESC
            LIMIT ?`, userID, tenantID, cursorTime, cursorTime, cursorID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*domain.Post, 0, limit)
	for rows.Next() {
		var p domain.Post
		if err := rows.Scan(&p.ID, &p.AuthorUserID, &p.Body, &p.CreatedAt, &p.LikeCount, &p.CommentCount, &p.LikedByMe); err != nil {
			return nil, err
		}
		items = append(items, &p)
	}
	return items, rows.Err()
}

func (r *timelineRepository) CreateComment(ctx context.Context, tenantID, postID, authorID uint64, body string) (*domain.Comment, error) {
	resExec, err := r.q.ExecContext(ctx, "INSERT INTO comments (tenant_id, post_id, author_user_id, body) VALUES (?,?,?,?)", tenantID, postID, authorID, body)
	if err != nil {
		return nil, err
	}
	id, _ := resExec.LastInsertId()
	var created time.Time
	_ = r.q.QueryRowContext(ctx, "SELECT created_at FROM comments WHERE id=?", id).Scan(&created)
	return &domain.Comment{
		ID:           uint64(id),
		PostID:       postID,
		AuthorUserID: authorID,
		Body:         body,
		CreatedAt:    created,
	}, nil
}

func (r *timelineRepository) FindCommentsByPostID(ctx context.Context, tenantID, postID uint64, limit int, cursorTime time.Time, cursorID uint64) ([]*domain.Comment, error) {
    var rows *sql.Rows
    var err error
    if cursorID == 0 {
        rows, err = r.q.QueryContext(ctx, `
            SELECT id, author_user_id, body, created_at
            FROM comments
            WHERE tenant_id=? AND post_id=?
            ORDER BY created_at ASC, id ASC
            LIMIT ?`, tenantID, postID, limit)
    } else {
        rows, err = r.q.QueryContext(ctx, `
            SELECT id, author_user_id, body, created_at
            FROM comments
            WHERE tenant_id=? AND post_id=? AND (created_at > ? OR (created_at = ? AND id > ?))
            ORDER BY created_at ASC, id ASC
            LIMIT ?`, tenantID, postID, cursorTime, cursorTime, cursorID, limit)
    }
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    items := make([]*domain.Comment, 0, limit)
    for rows.Next() {
        var cmt domain.Comment
        if err := rows.Scan(&cmt.ID, &cmt.AuthorUserID, &cmt.Body, &cmt.CreatedAt); err != nil {
            return nil, err
        }
        cmt.PostID = postID
        items = append(items, &cmt)
    }
    return items, rows.Err()
}
