package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/example/something-like-sns/apps/api/internal/port"
)

// DBTX is an interface that is satisfied by both *sql.DB and *sql.Tx
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// sqlStore provides all functions to execute db queries and transactions
type sqlStore struct {
	db *sql.DB
	q  DBTX
}

// NewStore creates a new Store
func NewStore(db *sql.DB) port.Store {
	return &sqlStore{
		db: db,
		q:  db,
	}
}

// ExecTx executes a function within a database transaction
func (s *sqlStore) ExecTx(ctx context.Context, fn func(port.Store) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txStore := &sqlStore{
		db: s.db,
		q:  tx,
	}

	err = fn(txStore)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

func (s *sqlStore) AuthRepository() port.AuthRepository {
	return &authRepository{q: s.q}
}

func (s *sqlStore) TimelineRepository() port.TimelineRepository {
	return &timelineRepository{q: s.q}
}

func (s *sqlStore) ReactionRepository() port.ReactionRepository {
	return &reactionRepository{q: s.q}
}

func (s *sqlStore) DMRepository() port.DMRepository {
	return &dmRepository{q: s.q}
}
