package train

import (
	"context"
	"database/sql"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

type Summary struct {
	ID, Number, Status string
	Capacity, Reserved int
	Departure          time.Time
}

func (s Service) List(ctx context.Context, u domain.User) ([]Summary, error) {
	rows, err := s.Store.Query(ctx, "SELECT id,number,status,capacity,reserved,departure_at FROM trains WHERE tenant_id=? ORDER BY departure_at,id", u.TenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Summary, 0)
	for rows.Next() {
		var x Summary
		var d string
		if err := rows.Scan(&x.ID, &x.Number, &x.Status, &x.Capacity, &x.Reserved, &d); err != nil {
			return nil, err
		}
		x.Departure, _ = time.Parse(time.RFC3339Nano, d)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s Service) Slot(ctx context.Context, u domain.User, id string) (string, error) {
	var status string
	err := s.Store.QueryRow(ctx, "SELECT status FROM rail_slots WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&status)
	if err == sql.ErrNoRows {
		return "", domain.ErrNotFound
	}
	return status, err
}
