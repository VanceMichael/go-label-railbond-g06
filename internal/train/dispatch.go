package train

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type DispatchResult struct {
	TrainID        string
	Moved, Skipped int
	At             time.Time
}

func (s Service) DispatchReady(ctx context.Context, u domain.User, id string) (DispatchResult, error) {
	result := DispatchResult{TrainID: id, At: time.Now().UTC()}
	err := s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		rows, err := tx.Query(ctx, "SELECT c.id,d.status FROM consignments c LEFT JOIN customs_declarations d ON d.consignment_id=c.id WHERE c.tenant_id=? AND c.train_id=? AND c.status='booked'", u.TenantID, id)
		if err != nil {
			return err
		}
		defer rows.Close()
		type x struct{ id, status string }
		items := []x{}
		for rows.Next() {
			var a x
			if err := rows.Scan(&a.id, &a.status); err != nil {
				return err
			}
			items = append(items, a)
		}
		for _, a := range items {
			if a.status != "released" {
				result.Skipped++
				continue
			}
			if _, err := tx.Exec(ctx, "UPDATE consignments SET status='in_transit',version=version+1 WHERE tenant_id=? AND id=? AND status='booked'", u.TenantID, a.id); err != nil {
				return err
			}
			result.Moved++
		}
		if result.Moved == 0 {
			return fmt.Errorf("%w: no customs-ready cargo", domain.ErrDeclarationHold)
		}
		return nil
	})
	return result, err
}
func (s Service) DepartAt(ctx context.Context, u domain.User, id string, at time.Time) error {
	if at.Before(time.Now().UTC()) {
		return fmt.Errorf("%w: departure is in the past", domain.ErrInvalidState)
	}
	return s.Depart(ctx, u, id)
}
