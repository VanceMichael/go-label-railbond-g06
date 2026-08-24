package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

func (s *Store) ResizeTrainCapacity(ctx context.Context, tenantID, id string, capacity int) error {
	return s.WithTx(ctx, func(tx *Tx) error {
		var reserved int
		if err := tx.QueryRow(ctx, "SELECT reserved FROM trains WHERE tenant_id=? AND id=?", tenantID, id).Scan(&reserved); err != nil {
			if IsMissing(err) {
				return fmt.Errorf("%w: train", sql.ErrNoRows)
			}
			return err
		}
		if capacity < reserved {
			return fmt.Errorf("%w: train capacity", domain.ErrConflict)
		}
		result, err := tx.Exec(ctx, "UPDATE trains SET capacity=?,version=version+1 WHERE tenant_id=? AND id=? AND reserved<=?", capacity, tenantID, id, capacity)
		if err != nil {
			return err
		}
		changed, _ := result.RowsAffected()
		if changed != 1 {
			return fmt.Errorf("%w: train capacity", domain.ErrConflict)
		}
		return nil
	})
}
