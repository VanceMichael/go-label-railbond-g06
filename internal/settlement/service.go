package settlement

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type Service struct{ Store *storage.Store }

func (s Service) SettleInvoice(ctx context.Context, u domain.User, invoiceID, providerKey, requestID string) error {
	if err := domain.RequireRole(u, "admin", "finance"); err != nil {
		return err
	}
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var status string
		var amount int
		var consignment string
		if err := tx.QueryRow(ctx, "SELECT status,amount,consignment_id FROM invoices WHERE tenant_id=? AND id=?", u.TenantID, invoiceID).Scan(&status, &amount, &consignment); err != nil {
			return err
		}
		if status == "disputed" {
			return fmt.Errorf("%w: disputed invoice", domain.ErrInvalidState)
		}
		if status != "issued" {
			return fmt.Errorf("%w: invoice status", domain.ErrInvalidState)
		}
		if s.Store.Hooks.BeforePayment != nil {
			if err := s.Store.Hooks.BeforePayment(ctx); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, "INSERT INTO payments(id,invoice_id,provider_key,status,amount,captured_at) VALUES(?,?,?,?,?,?)", storage.NewID(), invoiceID, providerKey, "captured", amount, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "UPDATE invoices SET status='settled',settled_at=? WHERE tenant_id=? AND id=? AND status='issued'", time.Now().UTC().Format(time.RFC3339Nano), u.TenantID, invoiceID); err != nil {
			return err
		}
		if err := s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "invoice.settled", "invoice", invoiceID, "success", requestID, fmt.Sprintf("amount=%d consignment=%s", amount, consignment)); err != nil {
			return err
		}
		return s.Store.Enqueue(ctx, tx, u.TenantID, "invoice.settled", invoiceID, providerKey)
	})
}
func (s Service) Dispute(ctx context.Context, u domain.User, id, reason string) error {
	if err := domain.RequireRole(u, "admin", "finance"); err != nil {
		return err
	}
	_, err := s.Store.Exec(ctx, "UPDATE invoices SET status='disputed',dispute_reason=? WHERE tenant_id=? AND id=? AND status='issued'", reason, u.TenantID, id)
	return err
}
func (s Service) Invoice(ctx context.Context, u domain.User, id string) (domain.InvoiceStatus, int, error) {
	var status string
	var amount int
	err := s.Store.QueryRow(ctx, "SELECT status,amount FROM invoices WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&status, &amount)
	return domain.InvoiceStatus(status), amount, err
}
