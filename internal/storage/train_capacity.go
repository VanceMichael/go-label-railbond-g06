package storage

import (
	"context"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

func (s *Store) ResizeTrainCapacity(ctx context.Context, tenantID, id string, capacity int) error {
	result, err := s.DB.ExecContext(ctx, "UPDATE trains SET capacity=?,version=version+1 WHERE tenant_id=? AND id=?", capacity, tenantID, id)
	if err != nil {
		return err
	}
	changed, _ := result.RowsAffected()
	if changed != 1 {
		return domain.ErrConflict
	}
	return nil
}
