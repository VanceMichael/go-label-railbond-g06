package warehouse

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type Occupancy struct{ Available, Reserved int }

func (s Service) Occupancy(ctx context.Context, u domain.User, zone string) (Occupancy, error) {
	var o Occupancy
	if err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM warehouse_slots WHERE tenant_id=? AND zone=? AND status='available'", u.TenantID, zone).Scan(&o.Available); err != nil {
		return o, err
	}
	if err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM warehouse_slots WHERE tenant_id=? AND zone=? AND status='reserved'", u.TenantID, zone).Scan(&o.Reserved); err != nil {
		return o, err
	}
	return o, nil
}
func (s Service) Move(ctx context.Context, u domain.User, from, to, consignmentID, requestID string) error {
	var oldStatus, newStatus string
	if err := s.Store.QueryRow(ctx, "SELECT status FROM warehouse_slots WHERE tenant_id=? AND id=?", u.TenantID, from).Scan(&oldStatus); err != nil {
		return err
	}
	if err := s.Store.QueryRow(ctx, "SELECT status FROM warehouse_slots WHERE tenant_id=? AND id=?", u.TenantID, to).Scan(&newStatus); err != nil {
		return err
	}
	if oldStatus != "reserved" || newStatus != "available" {
		return fmt.Errorf("%w: warehouse move", domain.ErrConflict)
	}
	if _, err := s.Store.Exec(ctx, "UPDATE warehouse_slots SET status='available',reserved_for=NULL,version=version+1 WHERE tenant_id=? AND id=? AND reserved_for=?", u.TenantID, from, consignmentID); err != nil {
		return err
	}
	if _, err := s.Store.Exec(ctx, "UPDATE warehouse_slots SET status='reserved',reserved_for=?,version=version+1 WHERE tenant_id=? AND id=? AND status='available'", consignmentID, u.TenantID, to); err != nil {
		return err
	}
	if _, err := s.Store.Exec(ctx, "UPDATE slot_reservations SET slot_id=? WHERE tenant_id=? AND consignment_id=? AND status='active'", to, u.TenantID, consignmentID); err != nil {
		return err
	}
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		return s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "warehouse.moved", "consignment", consignmentID, "success", requestID, from+"->"+to)
	})
}
func (s Service) Reservation(ctx context.Context, u domain.User, consignmentID string) (string, error) {
	var slot string
	err := s.Store.QueryRow(ctx, "SELECT slot_id FROM slot_reservations WHERE tenant_id=? AND consignment_id=? AND status='active'", u.TenantID, consignmentID).Scan(&slot)
	if err == sql.ErrNoRows {
		return "", domain.ErrNotFound
	}
	return slot, err
}
func (s Service) Age(ctx context.Context, u domain.User, consignmentID string) (time.Duration, error) {
	var at string
	if err := s.Store.QueryRow(ctx, "SELECT created_at FROM slot_reservations WHERE tenant_id=? AND consignment_id=? AND status='active'", u.TenantID, consignmentID).Scan(&at); err != nil {
		return 0, err
	}
	t, err := time.Parse(time.RFC3339Nano, at)
	if err != nil {
		return 0, err
	}
	return time.Since(t), nil
}
