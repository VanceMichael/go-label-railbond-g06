package schedule

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type Service struct{ Store *storage.Store }
type Window struct {
	ID         time.Time
	Start, End time.Time
	Status     string
}

func (s Service) ReserveWindow(ctx context.Context, u domain.User, corridor string, start, end time.Time) (string, error) {
	if !end.After(start) {
		return "", fmt.Errorf("%w: schedule window", domain.ErrInvalidState)
	}
	id := storage.NewID()
	err := s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var clashes int
		if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM rail_slots WHERE tenant_id=? AND corridor_id=? AND status IN ('held','used') AND starts_at<? AND ends_at>?", u.TenantID, corridor, end.Format(time.RFC3339Nano), start.Format(time.RFC3339Nano)).Scan(&clashes); err != nil {
			return err
		}
		if clashes > 0 {
			return domain.ErrConflict
		}
		_, err := tx.Exec(ctx, "INSERT INTO rail_slots(id,tenant_id,corridor_id,starts_at,ends_at,status,version) VALUES(?,?,?,?,?,?,1)", id, u.TenantID, corridor, start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano), "held")
		return err
	})
	return id, err
}
func (s Service) ReleaseWindow(ctx context.Context, u domain.User, id string) error {
	return s.Store.ReleaseRailWindow(ctx, u.TenantID, id)
}
func (s Service) IsOpen(start, end, timeNow time.Time) bool {
	return !timeNow.Before(start) && timeNow.Before(end)
}
