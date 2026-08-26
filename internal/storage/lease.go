package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

type LeaseRecord struct {
	ID, Owner, Token string
	Epoch            int
	Until            time.Time
}

func (s *Store) ClaimContainer(ctx context.Context, tenantID, containerID, owner, token string, until time.Time) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := s.DB.ExecContext(ctx, "UPDATE containers SET lease_owner=?,lease_token=?,lease_until=?,version=version+1 WHERE tenant_id=? AND id=? AND (lease_until IS NULL OR lease_until<?)", owner, token, until.UTC().Format(time.RFC3339Nano), tenantID, containerID, now)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return fmt.Errorf("%w: container lease", domain.ErrLeaseLost)
	}
	return nil
}

func (s *Store) RenewContainer(ctx context.Context, tenantID, id, owner, token string, until time.Time) error {
	res, err := s.DB.ExecContext(ctx, "UPDATE containers SET lease_until=?,version=version+1 WHERE tenant_id=? AND id=? AND lease_owner=? AND lease_token=? AND lease_until>?", until.UTC().Format(time.RFC3339Nano), tenantID, id, owner, token, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return fmt.Errorf("%w: container renewal", domain.ErrLeaseLost)
	}
	return nil
}

func (s *Store) ClaimOutbox(ctx context.Context, tenantID, id, owner string, epoch int, until time.Time) error {
	res, err := s.DB.ExecContext(ctx, "UPDATE outbox_messages SET status='sending',lease_owner=?,lease_epoch=?,lease_until=?,attempts=attempts+1 WHERE tenant_id=? AND id=? AND status='pending' AND (lease_until IS NULL OR lease_until<?)", owner, epoch, until.UTC().Format(time.RFC3339Nano), tenantID, id, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return fmt.Errorf("%w: outbox claim", domain.ErrLeaseLost)
	}
	return nil
}

func (s *Store) AckOutbox(ctx context.Context, tenantID, id, owner string, epoch int) error {
	res, err := s.DB.ExecContext(ctx, "DELETE FROM outbox_messages WHERE tenant_id=? AND id=? AND status='sending' AND lease_owner=? AND lease_epoch=?", tenantID, id, owner, epoch)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return fmt.Errorf("%w: outbox acknowledgement", domain.ErrLeaseLost)
	}
	return nil
}

func (s *Store) FailOutbox(ctx context.Context, tenantID, id, owner string, epoch int, reason string, availableAt time.Time) error {
	res, err := s.DB.ExecContext(ctx, "UPDATE outbox_messages SET status='pending',lease_owner=NULL,lease_until=NULL,last_error=?,available_at=? WHERE tenant_id=? AND id=? AND status='sending' AND lease_owner=? AND lease_epoch=?", reason, availableAt.UTC().Format(time.RFC3339Nano), tenantID, id, owner, epoch)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return fmt.Errorf("%w: outbox failure", domain.ErrLeaseLost)
	}
	return nil
}

func (s *Store) RowsAffected(res sql.Result) int64 { n, _ := res.RowsAffected(); return n }
