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
	return w.Store.ClaimRouteAssignment(ctx, tenantID, id, owner, epoch, until)
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
