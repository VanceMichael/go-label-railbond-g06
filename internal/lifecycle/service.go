package lifecycle

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type Service struct{ Store *storage.Store }

func (s Service) CanCloseTrain(ctx context.Context, u domain.User, id string) (bool, error) {
	var active, holds int
	if err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM consignments WHERE tenant_id=? AND train_id=? AND status NOT IN ('delivered','archived','draft')", u.TenantID, id).Scan(&active); err != nil {
		return false, err
	}
	if err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM customs_declarations d JOIN consignments c ON c.id=d.consignment_id WHERE c.tenant_id=? AND c.train_id=? AND d.status IN ('draft','submitted','hold')", u.TenantID, id).Scan(&holds); err != nil {
		return false, err
	}
	return active == 0 && holds == 0, nil
}
func (s Service) CloseTrain(ctx context.Context, u domain.User, id, requestID string) error {
	ok, err := s.CanCloseTrain(ctx, u, id)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: train has open cargo", domain.ErrInvalidState)
	}
	_, err = s.Store.Exec(ctx, "UPDATE trains SET status='arrived',version=version+1 WHERE tenant_id=? AND id=? AND status='departed'", u.TenantID, id)
	return err
}
