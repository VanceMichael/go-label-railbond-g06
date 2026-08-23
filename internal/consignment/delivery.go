package consignment

import (
	"context"
	"fmt"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type DeliveryService struct{ Store *storage.Store }

func (s DeliveryService) Deliver(ctx context.Context, u domain.User, id, signature, requestID string) (string, error) {
	if signature == "" {
		return "", fmt.Errorf("%w: receiver signature", domain.ErrInvalidState)
	}
	invoiceID := storage.NewID()
	err := s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var status string
		var train string
		if err := tx.QueryRow(ctx, "SELECT status,train_id FROM consignments WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&status, &train); err != nil {
			return err
		}
		if status != "in_transit" && status != "at_checkpoint" {
			return fmt.Errorf("%w: delivery state", domain.ErrInvalidState)
		}
		if _, err := tx.Exec(ctx, "UPDATE consignments SET status='delivered',delivered_at=?,version=version+1 WHERE tenant_id=? AND id=? AND status IN ('in_transit','at_checkpoint')", time.Now().UTC().Format(time.RFC3339Nano), u.TenantID, id); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "INSERT INTO invoices(id,tenant_id,consignment_id,status,amount,issued_at) VALUES(?,?,?,?,?,?)", invoiceID, u.TenantID, id, "issued", 0, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			return err
		}
		if err := s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "consignment.delivered", "consignment", id, "success", requestID, signature); err != nil {
			return err
		}
		return s.Store.Enqueue(ctx, tx, u.TenantID, "invoice.issued", invoiceID, id)
	})
	if err != nil {
		return "", domain.Wrap("deliver consignment", err)
	}
	return invoiceID, nil
}
