package train

import (
	"context"
	"fmt"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type CancellationService struct{ Store *storage.Store }

func (s CancellationService) Cancel(ctx context.Context, u domain.User, id string) error {
	if err := domain.RequireRole(u, "admin", "operator"); err != nil {
		return err
	}
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var slot string
		var status string
		if err := tx.QueryRow(ctx, "SELECT status,slot_id FROM trains WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&status, &slot); err != nil {
			return err
		}
		if !domain.TrainStatus(status).CanMove(domain.TrainCancelled) {
			return fmt.Errorf("%w: cancel train", domain.ErrInvalidState)
		}
		if _, err := tx.Exec(ctx, "UPDATE consignments SET status='draft',version=version+1 WHERE tenant_id=? AND train_id=? AND status IN ('booked','in_transit')", u.TenantID, id); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "UPDATE trains SET status='cancelled',version=version+1 WHERE tenant_id=? AND id=? AND status IN ('planned','published')", u.TenantID, id); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "UPDATE rail_slots SET status='available',train_id=NULL,version=version+1 WHERE tenant_id=? AND id=? AND status='held'", u.TenantID, slot); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, "UPDATE slot_reservations SET status='released',released_at=datetime('now') WHERE tenant_id=? AND consignment_id IN (SELECT id FROM consignments WHERE train_id=?) AND status='active'", u.TenantID, id)
		return err
	})
}
