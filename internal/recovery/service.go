package recovery

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type Service struct{ Store *storage.Store }
type Report struct{ OutboxReset, AssignmentsReset, ContainersReleased int }

func (s Service) RecoverExpired(ctx context.Context, now time.Time) (Report, error) {
	var report Report
	reset, err := s.Store.ResetExpiredOutbox(ctx, now)
	if err != nil {
		return report, err
	}
	report.OutboxReset = reset
	err = s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		result, err := tx.Exec(ctx, "UPDATE route_assignments SET status='retry',lease_owner=NULL,lease_until=NULL,last_error='lease expired' WHERE status='running' AND lease_until<?", now.Format(time.RFC3339Nano))
		if err != nil {
			return err
		}
		changed, _ := result.RowsAffected()
		report.AssignmentsReset = int(changed)
		result, err = tx.Exec(ctx, "UPDATE containers SET lease_owner=NULL,lease_token=NULL,lease_until=NULL,version=version+1 WHERE lease_until<?", now.Format(time.RFC3339Nano))
		if err != nil {
			return err
		}
		changed, _ = result.RowsAffected()
		report.ContainersReleased = int(changed)
		return nil
	})
	return report, err
}
func (s Service) MarkPermanentFailure(ctx context.Context, tenantID, table, id, reason string) error {
	if table != "outbox_messages" && table != "route_assignments" {
		return fmt.Errorf("%w: recovery target", domain.ErrInvalidState)
	}
	_, err := s.Store.Exec(ctx, "UPDATE "+table+" SET status='dead',last_error=? WHERE tenant_id=? AND id=?", reason, tenantID, id)
	return err
}
