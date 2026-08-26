package dispatch

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type BatchDispatcher struct{ Store *storage.Store }

func (s BatchDispatcher) Dispatch(ctx context.Context, u domain.User, trainID, requestID string) (int, error) {
	moved := 0
	err := s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		rows, err := tx.Query(ctx, "SELECT id,status FROM consignments WHERE tenant_id=? AND train_id=? ORDER BY id", u.TenantID, trainID)
		if err != nil {
			return err
		}
		defer rows.Close()
		type item struct{ id, status string }
		items := []item{}
		for rows.Next() {
			var x item
			if err := rows.Scan(&x.id, &x.status); err != nil {
				return err
			}
			items = append(items, x)
		}
		for _, x := range items {
			if x.status != "booked" {
				return fmt.Errorf("%w: batch consignment %s", domain.ErrInvalidState, x.id)
			}
		}
		for _, x := range items {
			if _, err := tx.Exec(ctx, "UPDATE consignments SET status='in_transit',version=version+1 WHERE tenant_id=? AND id=? AND status='booked'", u.TenantID, x.id); err != nil {
				return err
			}
			moved++
		}
		if err := s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "dispatch.batch", "train", trainID, "success", requestID, fmt.Sprintf("moved=%d", moved)); err != nil {
			return err
		}
		return nil
	})
	return moved, err
}
