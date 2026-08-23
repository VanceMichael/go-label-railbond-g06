package dispatch

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type LeaseService struct{ Store *storage.Store }

func (s LeaseService) ClaimContainer(ctx context.Context, u domain.User, id, owner, token string, until time.Time) error {
	return s.Store.ClaimContainer(ctx, u.TenantID, id, owner, token, until)
}
func (s LeaseService) RenewContainer(ctx context.Context, u domain.User, id, owner, token string, until time.Time) error {
	return s.Store.RenewContainer(ctx, u.TenantID, id, owner, token, until)
}
func (s LeaseService) ReleaseContainer(ctx context.Context, u domain.User, id, owner, token string) error {
	res, err := s.Store.Exec(ctx, "UPDATE containers SET lease_owner=NULL,lease_token=NULL,lease_until=NULL,version=version+1 WHERE tenant_id=? AND id=? AND lease_owner=? AND lease_token=?", u.TenantID, id, owner, token)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return fmt.Errorf("%w: release container", domain.ErrLeaseLost)
	}
	return nil
}
