package train

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"time"
)

type Health struct {
	TrainID, Status                              string
	Capacity, Reserved, Manifest, PendingCustoms int
	Departure                                    time.Time
}

func (s Service) Health(ctx context.Context, u domain.User, id string) (Health, error) {
	var h Health
	var departure string
	if err := s.Store.QueryRow(ctx, "SELECT id,status,capacity,reserved,departure_at FROM trains WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&h.TrainID, &h.Status, &h.Capacity, &h.Reserved, &departure); err != nil {
		if err == sql.ErrNoRows {
			return h, domain.ErrNotFound
		}
		return h, err
	}
	h.Departure, _ = time.Parse(time.RFC3339Nano, departure)
	if err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM consignments WHERE tenant_id=? AND train_id=? AND status NOT IN ('draft','archived')", u.TenantID, id).Scan(&h.Manifest); err != nil {
		return h, err
	}
	if err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM customs_declarations d JOIN consignments c ON c.id=d.consignment_id WHERE c.tenant_id=? AND c.train_id=? AND d.status!='released'", u.TenantID, id).Scan(&h.PendingCustoms); err != nil {
		return h, err
	}
	if h.Reserved > h.Capacity {
		return h, fmt.Errorf("%w: train capacity", domain.ErrConflict)
	}
	return h, nil
}
func (s Service) ReadyForDeparture(ctx context.Context, u domain.User, id string) (bool, error) {
	h, err := s.Health(ctx, u, id)
	if err != nil {
		return false, err
	}
	return h.Status == "published" && h.PendingCustoms == 0 && h.Manifest > 0 && h.Reserved <= h.Capacity, nil
}
func (s Service) ScheduleSummary(ctx context.Context, u domain.User, from, to time.Time) ([]Health, error) {
	rows, err := s.Store.Query(ctx, "SELECT id FROM trains WHERE tenant_id=? AND departure_at>=? AND departure_at<? ORDER BY departure_at,id", u.TenantID, from.Format(time.RFC3339Nano), to.Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Health{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		h, err := s.Health(ctx, u, id)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}
