package customs

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type BatchService struct{ Store *storage.Store }

func (s BatchService) Release(ctx context.Context, u domain.User, ids []string, requestID string) (int, error) {
	if len(ids) == 0 {
		return 0, domain.ErrInvalidState
	}
	released := 0
	for _, id := range ids {
		err := s.Store.WithTx(ctx, func(tx *storage.Tx) error {
			var status string
			if err := tx.QueryRow(ctx, "SELECT status FROM customs_declarations WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&status); err != nil {
				return err
			}
			if status == "hold" {
				return fmt.Errorf("%w: batch item %s", domain.ErrDeclarationHold, id)
			}
			if status != "submitted" {
				return fmt.Errorf("%w: batch item %s", domain.ErrInvalidState, id)
			}
			if _, err := tx.Exec(ctx, "UPDATE customs_declarations SET status='released',released_at=datetime('now'),version=version+1 WHERE tenant_id=? AND id=? AND status='submitted'", u.TenantID, id); err != nil {
				return err
			}
			return s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "customs.batch_item_released", "declaration", id, "success", requestID, "partial batch")
		})
		if err != nil {
			return released, err
		}
		released++
	}
	return released, nil
}
