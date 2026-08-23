package warehouse

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type ZoneMetric struct {
	Zone                      string
	Available, Reserved, Held int
	OldestReservation         time.Time
}

func (s Service) Metrics(ctx context.Context, u domain.User) ([]ZoneMetric, error) {
	rows, err := s.Store.Query(ctx, "SELECT zone, SUM(CASE WHEN status='available' THEN 1 ELSE 0 END), SUM(CASE WHEN status='reserved' THEN 1 ELSE 0 END), SUM(CASE WHEN status='held' THEN 1 ELSE 0 END) FROM warehouse_slots WHERE tenant_id=? GROUP BY zone ORDER BY zone", u.TenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ZoneMetric{}
	for rows.Next() {
		var m ZoneMetric
		if err := rows.Scan(&m.Zone, &m.Available, &m.Reserved, &m.Held); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (s Service) AssertCapacity(ctx context.Context, u domain.User, zone string, needed int) error {
	var available int
	if err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM warehouse_slots WHERE tenant_id=? AND zone=? AND status='available'", u.TenantID, zone).Scan(&available); err != nil {
		return err
	}
	if available < needed {
		return fmt.Errorf("%w: zone %s needs %d slots but has %d", domain.ErrConflict, zone, needed, available)
	}
	return nil
}
func (s Service) ReservationsBefore(ctx context.Context, u domain.User, before time.Time) ([]string, error) {
	rows, err := s.Store.Query(ctx, "SELECT consignment_id FROM slot_reservations WHERE tenant_id=? AND status='active' AND created_at<? ORDER BY created_at,consignment_id", u.TenantID, before.Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
func (s Service) ReservationExists(ctx context.Context, u domain.User, id string) (bool, error) {
	var n int
	err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM slot_reservations WHERE tenant_id=? AND consignment_id=? AND status IN ('active','held')", u.TenantID, id).Scan(&n)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return n > 0, err
}
func (s Service) ReconcileReservations(ctx context.Context, u domain.User) error {
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		rows, err := tx.Query(ctx, "SELECT slot_id,consignment_id FROM slot_reservations WHERE tenant_id=? AND status='active'", u.TenantID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var slot, consignment string
			if err := rows.Scan(&slot, &consignment); err != nil {
				return err
			}
			var reserved sql.NullString
			if err := tx.QueryRow(ctx, "SELECT reserved_for FROM warehouse_slots WHERE tenant_id=? AND id=?", u.TenantID, slot).Scan(&reserved); err != nil {
				return err
			}
			if !reserved.Valid || reserved.String != consignment {
				return fmt.Errorf("%w: reservation %s", domain.ErrConflict, consignment)
			}
		}
		return nil
	})
}
