package query

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type ETAService struct{ Store *storage.Store }

func (s ETAService) Estimate(ctx context.Context, u domain.User, id string) (time.Time, error) {
	var departure string
	var last string
	if err := domain.CheckContext(ctx); err != nil {
		return time.Time{}, err
	}
	if err := s.Store.QueryRow(ctx, "SELECT t.departure_at FROM consignments c JOIN trains t ON t.id=c.train_id WHERE c.tenant_id=? AND c.id=?", u.TenantID, id).Scan(&departure); err != nil {
		return time.Time{}, err
	}
	if err := domain.CheckContext(ctx); err != nil {
		return time.Time{}, err
	}
	if err := s.Store.QueryRow(ctx, "SELECT COALESCE(MAX(observed_at),'') FROM checkpoint_events WHERE tenant_id=? AND consignment_id=?", u.TenantID, id).Scan(&last); err != nil {
		return time.Time{}, err
	}
	d, e := time.Parse(time.RFC3339Nano, departure)
	if e != nil {
		return time.Time{}, e
	}
	if last != "" {
		_, _ = time.Parse(time.RFC3339Nano, last)
		d = d.Add(time.Hour)
	}
	return d, nil
}
func (s ETAService) Delay(ctx context.Context, u domain.User, id string) (time.Duration, error) {
	eta, err := s.Estimate(ctx, u, id)
	if err != nil {
		return 0, fmt.Errorf("delay: %w", err)
	}
	return time.Until(eta), nil
}
