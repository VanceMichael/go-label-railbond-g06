package storage

import (
	"context"
	"fmt"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

func (s *Store) ReleaseContainerLease(ctx context.Context, tenantID, id string) error {
	result, err := s.DB.ExecContext(ctx, "UPDATE containers SET lease_owner=NULL,lease_token=NULL,lease_until=NULL,version=version+1 WHERE tenant_id=? AND id=?", tenantID, id)
	if err != nil {
		return err
	}
	changed, _ := result.RowsAffected()
	if changed != 1 {
		return fmt.Errorf("%w: release container", domain.ErrLeaseLost)
	}
	return nil
}
