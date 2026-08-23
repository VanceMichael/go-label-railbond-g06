package worker

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type AssignmentWorker struct{ Store *storage.Store }

func (w AssignmentWorker) Claim(ctx context.Context, tenantID, id, owner string, epoch int, until time.Time) error {
	res, err := w.Store.Exec(ctx, "UPDATE route_assignments SET status='running',lease_owner=?,lease_epoch=?,lease_until=?,attempt=attempt+1 WHERE tenant_id=? AND id=? AND status IN ('assigned','retry')", owner, epoch, until.UTC().Format(time.RFC3339Nano), tenantID, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return domain.ErrLeaseLost
	}
	return nil
}
func (w AssignmentWorker) Complete(ctx context.Context, tenantID, id, owner string, epoch int) error {
	res, err := w.Store.Exec(ctx, "UPDATE route_assignments SET status='delivered',lease_owner=NULL,lease_until=NULL WHERE tenant_id=? AND id=? AND status='running' AND lease_owner=? AND lease_epoch=?", tenantID, id, owner, epoch)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return fmt.Errorf("%w: assignment completion", domain.ErrLeaseLost)
	}
	return nil
}
