package warehouse

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type Service struct{ Store *storage.Store }

func (s Service) Reserve(ctx context.Context, u domain.User, slotID, consignmentID, requestID string) error {
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var status string
		if err := tx.QueryRow(ctx, "SELECT status FROM warehouse_slots WHERE tenant_id=? AND id=?", u.TenantID, slotID).Scan(&status); err != nil {
			return err
		}
		if status != "available" {
			return fmt.Errorf("%w: warehouse slot", domain.ErrConflict)
		}
		if _, err := tx.Exec(ctx, "UPDATE warehouse_slots SET status='reserved',reserved_for=?,version=version+1 WHERE tenant_id=? AND id=? AND status='available'", consignmentID, u.TenantID, slotID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "INSERT INTO slot_reservations(id,tenant_id,slot_id,consignment_id,status,created_at) VALUES(?,?,?,?,?,?)", storage.NewID(), u.TenantID, slotID, consignmentID, "active", time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			return err
		}
		return s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "warehouse.reserved", "slot", slotID, "success", requestID, consignmentID)
	})
}
func (s Service) ReleaseForCustomsRejection(ctx context.Context, u domain.User, consignmentID, requestID string) error {
	if s.Store.Hooks.BeforeWarehouseRelease != nil {
		if err := s.Store.Hooks.BeforeWarehouseRelease(ctx); err != nil {
			return err
		}
	}
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		if _, err := tx.Exec(ctx, "UPDATE slot_reservations SET status='released',released_at=? WHERE tenant_id=? AND consignment_id=? AND status='active'", time.Now().UTC().Format(time.RFC3339Nano), u.TenantID, consignmentID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "UPDATE warehouse_slots SET status='available',reserved_for=NULL,version=version+1 WHERE tenant_id=? AND reserved_for=?", u.TenantID, consignmentID); err != nil {
			return err
		}
		// Record the audit event inside the same transaction so a failure
		// rolls back the slot/reservation release atomically. Recording it
		// after commit would leave the warehouse positions freed while the
		// caller sees an error, allowing other cargo to claim the slot.
		return s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "warehouse.released", "consignment", consignmentID, "success", requestID, "customs rejection")
	})
}
func (s Service) List(ctx context.Context, u domain.User) ([]string, error) {
	rows, err := s.Store.Query(ctx, "SELECT code FROM warehouse_slots WHERE tenant_id=? AND status='available' ORDER BY code", u.TenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var x string
		if err := rows.Scan(&x); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
