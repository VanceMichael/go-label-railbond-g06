package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Hooks struct {
	BeforeAudit            func(context.Context) error
	BeforeOutbox           func(context.Context) error
	BeforePayment          func(context.Context) error
	BeforeCheckpoint       func(context.Context) error
	BeforeWarehouseRelease func(context.Context) error
	BeforeRouteCall        func(context.Context) error
	// BeforeRouteRecovery runs after the expired-outbox reset and before the
	// route-assignment reset inside RecoverExpired. Returning an error rolls the
	// whole recovery batch (outbox reset included) back, preserving the
	// all-or-nothing guarantee of the operation.
	BeforeRouteRecovery func(context.Context) error
}

type Store struct {
	DB    *sql.DB
	Hooks Hooks
	mu    sync.RWMutex
}

func Open(path string) (*Store, error) {
	if path == "" {
		path = ":memory:"
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	s := &Store{DB: db}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error { return s.DB.Close() }

func (s *Store) Migrate(ctx context.Context, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}
	_, err = s.DB.ExecContext(ctx, string(data))
	if err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}
	return nil
}

type Tx struct {
	tx    *sql.Tx
	store *Store
}

func (s *Store) WithTx(ctx context.Context, fn func(*Tx) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	wrapper := &Tx{tx: tx, store: s}
	if err := fn(wrapper); err != nil {
		if rb := tx.Rollback(); rb != nil {
			return fmt.Errorf("%w; rollback: %v", err, rb)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (t *Tx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}
func (t *Tx) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}
func (t *Tx) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}
func (t *Tx) Store() *Store { return t.store }

func (s *Store) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.DB.ExecContext(ctx, query, args...)
}
func (s *Store) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return s.DB.QueryRowContext(ctx, query, args...)
}
func (s *Store) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.DB.QueryContext(ctx, query, args...)
}

func (s *Store) RecordAudit(ctx context.Context, tx *Tx, tenantID, actorID, action, objectType, objectID, outcome, requestID, detail string) error {
	if s.Hooks.BeforeAudit != nil {
		if err := s.Hooks.BeforeAudit(ctx); err != nil {
			return fmt.Errorf("audit hook: %w", err)
		}
	}
	_, err := tx.Exec(ctx, "INSERT INTO audit_events(id,tenant_id,actor_id,action,object_type,object_id,outcome,request_id,detail,created_at) VALUES(?,?,?,?,?,?,?,?,?,?)",
		NewID(), tenantID, actorID, action, objectType, objectID, outcome, requestID, detail, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

func (s *Store) Enqueue(ctx context.Context, tx *Tx, tenantID, topic, aggregateID, payload string) error {
	if s.Hooks.BeforeOutbox != nil {
		if err := s.Hooks.BeforeOutbox(ctx); err != nil {
			return fmt.Errorf("outbox hook: %w", err)
		}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := tx.Exec(ctx, "INSERT INTO outbox_messages(id,tenant_id,topic,aggregate_id,payload,status,available_at,created_at) VALUES(?,?,?,?,?,?,?,?)",
		NewID(), tenantID, topic, aggregateID, payload, "pending", now, now)
	return err
}

func (s *Store) IsClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return false
}

func NewID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), strings.ReplaceAll(fmt.Sprintf("%p", &struct{}{}), "0x", ""))
}

func IsMissing(err error) bool { return errors.Is(err, sql.ErrNoRows) }
