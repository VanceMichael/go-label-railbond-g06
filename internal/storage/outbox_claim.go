package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

func (s *Store) ClaimOutboxWithoutLeaseGuard(ctx context.Context, tenantID, id, owner string, epoch int, until time.Time) error {
	result, err := s.DB.ExecContext(ctx, "UPDATE outbox_messages SET status='sending',lease_owner=?,lease_epoch=?,lease_until=?,attempts=attempts+1 WHERE tenant_id=? AND id=? AND status='pending'", owner, epoch, until.UTC().Format(time.RFC3339Nano), tenantID, id)
	if err != nil {
		return err
	}
	changed, _ := result.RowsAffected()
	if changed != 1 {
		return fmt.Errorf("%w: outbox claim", domain.ErrLeaseLost)
	}
	return nil
}
