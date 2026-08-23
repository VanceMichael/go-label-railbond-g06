package warehouse

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type Hold struct{ SlotID, ConsignmentID, Reason string }

func (s Service) PlaceHold(ctx context.Context, u domain.User, slotID, reason string) error {
	if reason == "" {
		return fmt.Errorf("%w: warehouse hold reason", domain.ErrInvalidState)
	}
	res, err := s.Store.Exec(ctx, "UPDATE warehouse_slots SET status='held',version=version+1 WHERE tenant_id=? AND id=? AND status='available'", u.TenantID, slotID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}
func (s Service) ReleaseHold(ctx context.Context, u domain.User, slotID string) error {
	res, err := s.Store.Exec(ctx, "UPDATE warehouse_slots SET status='available',version=version+1 WHERE tenant_id=? AND id=? AND status='held'", u.TenantID, slotID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}
func (s Service) HoldForConsignment(ctx context.Context, u domain.User, slotID, consignmentID, reason string) error {
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var status string
		if err := tx.QueryRow(ctx, "SELECT status FROM warehouse_slots WHERE tenant_id=? AND id=?", u.TenantID, slotID).Scan(&status); err != nil {
			return err
		}
		if status != "available" {
			return domain.ErrConflict
		}
		if _, err := tx.Exec(ctx, "UPDATE warehouse_slots SET status='held',reserved_for=?,version=version+1 WHERE tenant_id=? AND id=? AND status='available'", consignmentID, u.TenantID, slotID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, "INSERT INTO slot_reservations(id,tenant_id,slot_id,consignment_id,status,created_at) VALUES(?,?,?,?,?,datetime('now'))", storage.NewID(), u.TenantID, slotID, consignmentID, "held")
		return err
	})
}
