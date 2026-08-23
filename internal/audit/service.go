package audit

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type Event struct{ ID, Action, ObjectType, ObjectID, Outcome, RequestID, Detail, CreatedAt string }
type Service struct{ Store *storage.Store }

func (s Service) Write(ctx context.Context, u domain.User, action, objectType, objectID, outcome, requestID, detail string) error {
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		return s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, action, objectType, objectID, outcome, requestID, detail)
	})
}
func (s Service) Recent(ctx context.Context, u domain.User, limit int) ([]Event, error) {
	if limit < 1 {
		limit = 20
	}
	rows, err := s.Store.Query(ctx, "SELECT id,action,object_type,object_id,outcome,request_id,detail,created_at FROM audit_events WHERE tenant_id=? ORDER BY created_at DESC,id DESC LIMIT ?", u.TenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Event{}
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Action, &e.ObjectType, &e.ObjectID, &e.Outcome, &e.RequestID, &e.Detail, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
func (s Service) ForObject(ctx context.Context, u domain.User, kind, id string) ([]Event, error) {
	rows, err := s.Store.Query(ctx, "SELECT id,action,object_type,object_id,outcome,request_id,detail,created_at FROM audit_events WHERE tenant_id=? AND object_type=? AND object_id=? ORDER BY created_at", u.TenantID, kind, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Event{}
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Action, &e.ObjectType, &e.ObjectID, &e.Outcome, &e.RequestID, &e.Detail, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
func (s Service) VerifyFresh(ctx context.Context, u domain.User, id string, within time.Duration) error {
	var created string
	if err := s.Store.QueryRow(ctx, "SELECT created_at FROM audit_events WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&created); err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrNotFound
		}
		return err
	}
	t, err := time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return fmt.Errorf("audit timestamp: %w", err)
	}
	if time.Since(t) > within {
		return domain.ErrConflict
	}
	return nil
}
