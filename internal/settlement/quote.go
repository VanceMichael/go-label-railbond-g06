package settlement

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"math"
	"time"
)

type QuoteLine struct {
	Code                        string
	Quantity, UnitPrice, Amount int
}
type Quote struct {
	ConsignmentID        string
	Lines                []QuoteLine
	Subtotal, Tax, Total int
	ExpiresAt            time.Time
}

func (s Service) Quote(ctx context.Context, u domain.User, id string, expires time.Time) (Quote, error) {
	rows, err := s.Store.Query(ctx, "SELECT sku,quantity,declared_value FROM consignment_items WHERE consignment_id=? ORDER BY sku", id)
	if err != nil {
		return Quote{}, err
	}
	defer rows.Close()
	q := Quote{ConsignmentID: id, ExpiresAt: expires, Lines: []QuoteLine{}}
	for rows.Next() {
		var code string
		var qty, price int
		if err := rows.Scan(&code, &qty, &price); err != nil {
			return q, err
		}
		line := QuoteLine{Code: code, Quantity: qty, UnitPrice: price, Amount: qty * price}
		q.Lines = append(q.Lines, line)
		q.Subtotal += line.Amount
	}
	q.Tax = int(math.Round(float64(q.Subtotal) * 0.09))
	q.Total = q.Subtotal + q.Tax
	if len(q.Lines) == 0 {
		return q, fmt.Errorf("%w: empty quote", domain.ErrInvalidState)
	}
	return q, nil
}
func (s Service) PersistQuote(ctx context.Context, u domain.User, id string, q Quote, requestID string) error {
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var status string
		if err := tx.QueryRow(ctx, "SELECT status FROM consignments WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&status); err != nil {
			return err
		}
		if status != "delivered" && status != "at_checkpoint" {
			return fmt.Errorf("%w: quote state", domain.ErrInvalidState)
		}
		_, err := tx.Exec(ctx, "UPDATE consignments SET version=version+1 WHERE tenant_id=? AND id=?", u.TenantID, id)
		if err != nil {
			return err
		}
		return s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "settlement.quoted", "consignment", id, "success", requestID, fmt.Sprintf("total=%d expires=%s", q.Total, q.ExpiresAt.Format(time.RFC3339Nano)))
	})
}
