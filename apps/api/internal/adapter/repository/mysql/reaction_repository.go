package mysql

import (
	"context"
	"database/sql"

	"github.com/example/something-like-sns/apps/api/internal/domain"
)

type reactionRepository struct {
	db *sql.DB
}

func NewReactionRepository(db *sql.DB) *reactionRepository {
	return &reactionRepository{db: db}
}

func (r *reactionRepository) Toggle(ctx context.Context, tenantID, userID uint64, targetType domain.ReactionTargetType, targetID uint64, reactionType string) (bool, error) {
	res, err := r.db.ExecContext(ctx, "DELETE FROM reactions WHERE tenant_id=? AND target_type=? AND target_id=? AND user_id=? AND type=?", tenantID, targetType, targetID, userID, reactionType)
	if err != nil {
		return false, err
	}
	affected, _ := res.RowsAffected()
	active := false
	if affected == 0 {
		if _, err := r.db.ExecContext(ctx, "INSERT INTO reactions (tenant_id, target_type, target_id, user_id, type) VALUES (?,?,?,?,?)", tenantID, targetType, targetID, userID, reactionType); err != nil {
			return false, err
		}
		active = true
	}
	return active, nil
}

func (r *reactionRepository) Count(ctx context.Context, tenantID uint64, targetType domain.ReactionTargetType, targetID uint64, reactionType string) (uint32, error) {
	var total uint32
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM reactions WHERE tenant_id=? AND target_type=? AND target_id=? AND type=?", tenantID, targetType, targetID, reactionType).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}
