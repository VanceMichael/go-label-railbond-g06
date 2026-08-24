package dispatch

import (
	"context"
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
	return s.Store.ReleaseContainerLease(ctx, u.TenantID, id, owner, token)
}
