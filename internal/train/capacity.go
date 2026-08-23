package train

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

type CapacityReport struct{ Capacity, Reserved, Available int }

func (s Service) Capacity(ctx context.Context, u domain.User, id string) (CapacityReport, error) {
	var r CapacityReport
	if err := s.Store.QueryRow(ctx, "SELECT capacity,reserved,capacity-reserved FROM trains WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&r.Capacity, &r.Reserved, &r.Available); err != nil {
		return r, err
	}
	if r.Available < 0 {
		return r, fmt.Errorf("%w: negative capacity", domain.ErrConflict)
	}
	return r, nil
}
func (s Service) Resize(ctx context.Context, u domain.User, id string, newCapacity int) error {
	if err := domain.RequireRole(u, "admin"); err != nil {
		return err
	}
	return s.Store.ResizeTrainCapacity(ctx, u.TenantID, id, newCapacity)
}
func (s Service) ReserveForConsignment(ctx context.Context, u domain.User, id string) error {
	return s.ReserveCapacity(ctx, u, id, 1)
}
