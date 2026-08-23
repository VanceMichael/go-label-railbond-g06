package query

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type OverviewService struct{ Store *storage.Store }
type Overview struct{ Consignments, Released, OpenExceptions int }

func (s OverviewService) Load(ctx context.Context, u domain.User) (Overview, error) {
	var o Overview
	err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM consignments WHERE tenant_id=?", u.TenantID).Scan(&o.Consignments)
	if err != nil {
		return o, err
	}
	if err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM customs_declarations WHERE tenant_id=? AND status='released'", u.TenantID).Scan(&o.Released); err != nil {
		return o, err
	}
	err = s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM exceptions WHERE tenant_id=? AND status='open'", u.TenantID).Scan(&o.OpenExceptions)
	return o, err
}
