package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type TransactionLog struct {
	StartedAt, FinishedAt time.Time
	Committed             bool
}

func (s *Store) Ping(ctx context.Context) error { return s.DB.PingContext(ctx) }
func (s *Store) WithReadTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("begin read transaction: %w", err)
	}
	if err := fn(tx); err != nil {
		if rb := tx.Rollback(); rb != nil {
			return fmt.Errorf("%w; rollback: %v", err, rb)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit read transaction: %w", err)
	}
	return nil
}
func (s *Store) Vacuum(ctx context.Context) error {
	_, err := s.DB.ExecContext(ctx, "PRAGMA optimize")
	return err
}
func (s *Store) ForeignKeys(ctx context.Context) (bool, error) {
	var v int
	if err := s.DB.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&v); err != nil {
		return false, err
	}
	return v == 1, nil
}
