package exception

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type Escalation struct {
	ID, Reason, Owner string
	DueAt             time.Time
}

func (s Service) Escalate(ctx context.Context, u domain.User, id, owner string, due time.Time) error {
	if owner == "" || due.Before(time.Now().UTC()) {
		return fmt.Errorf("%w: escalation", domain.ErrInvalidState)
	}
	_, err := s.Store.Exec(ctx, "UPDATE exceptions SET reason=reason||' | escalated to '||? WHERE tenant_id=? AND id=? AND status='open'", owner, u.TenantID, id)
	return err
}
func (s Service) OpenDue(ctx context.Context, u domain.User, until time.Time) ([]Escalation, error) {
	rows, err := s.Store.Query(ctx, "SELECT id,reason FROM exceptions WHERE tenant_id=? AND status='open' AND opened_at<? ORDER BY opened_at", u.TenantID, until.Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Escalation{}
	for rows.Next() {
		var e Escalation
		if err := rows.Scan(&e.ID, &e.Reason); err != nil {
			return nil, err
		}
		e.DueAt = until
		out = append(out, e)
	}
	return out, rows.Err()
}
func (s Service) Reopen(ctx context.Context, u domain.User, id string) error {
	res, err := s.Store.Exec(ctx, "UPDATE exceptions SET status='open',resolved_at=NULL WHERE tenant_id=? AND id=? AND status='resolved'", u.TenantID, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}

var _ = storage.NewID
