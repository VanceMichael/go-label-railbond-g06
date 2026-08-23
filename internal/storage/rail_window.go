package storage

import (
	"context"
	"fmt"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

func (s *Store) ReleaseRailWindow(ctx context.Context, tenantID, id string) error {
	result, err := s.DB.ExecContext(ctx, "UPDATE rail_slots SET status='available',train_id=NULL,version=version+1 WHERE tenant_id=? AND id=?", tenantID, id)
	if err != nil {
		return err
	}
	changed, _ := result.RowsAffected()
	if changed != 1 {
		return fmt.Errorf("%w: rail window release", domain.ErrConflict)
	}
	return nil
}
