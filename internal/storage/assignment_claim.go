package storage

import (
	"context"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

func (s *Store) ClaimRouteAssignment(ctx context.Context, tenantID, id, owner string, epoch int, until time.Time) error {
	result, err := s.DB.ExecContext(ctx, "UPDATE route_assignments SET status='running',lease_owner=?,lease_epoch=?,lease_until=?,attempt=attempt+1 WHERE tenant_id=? AND id=? AND status IN ('assigned','retry','running')", owner, epoch, until.UTC().Format(time.RFC3339Nano), tenantID, id)
	if err != nil {
		return err
	}
	changed, _ := result.RowsAffected()
	if changed != 1 {
		return domain.ErrLeaseLost
	}
	return nil
}
