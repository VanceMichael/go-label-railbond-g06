package worker

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"net/http"
	"time"
)

type RouteAssignmentWorker struct {
	Store  *storage.Store
	Client *http.Client
}

func (w RouteAssignmentWorker) Claim(ctx context.Context, tenantID, id, owner string, epoch int, until time.Time) error {
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

func (w RouteAssignmentWorker) RunOnce(ctx context.Context, tenantID, id, owner string, epoch int) error {
	if err := w.Claim(ctx, tenantID, id, owner, epoch, time.Now().UTC().Add(time.Minute)); err != nil {
		return err
	}
	if w.Store.Hooks.BeforeRouteCall != nil {
		if err := w.Store.Hooks.BeforeRouteCall(ctx); err != nil {
			_, _ = w.Store.Exec(context.Background(), "UPDATE route_assignments SET status='retry',lease_owner=NULL,lease_until=NULL,last_error=? WHERE tenant_id=? AND id=? AND lease_owner=?", err.Error(), tenantID, id, owner)
			return err
		}
	}
	if err := domain.CheckContext(ctx); err != nil {
		_, _ = w.Store.Exec(context.Background(), "UPDATE route_assignments SET status='retry',lease_owner=NULL,lease_until=NULL,last_error=? WHERE tenant_id=? AND id=? AND lease_owner=?", err.Error(), tenantID, id, owner)
		return err
	}
	if _, err := w.Store.Exec(ctx, "UPDATE route_assignments SET status='delivered',lease_owner=NULL,lease_until=NULL WHERE tenant_id=? AND id=? AND lease_owner=? AND lease_epoch=?", tenantID, id, owner, epoch); err != nil {
		return fmt.Errorf("complete route: %w", err)
	}
	return nil
}
