package settlement

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type Reconciliation struct {
	InvoiceID, PaymentID, InvoiceStatus, PaymentStatus string
	Amount                                             int
}

func (s Service) Reconcile(ctx context.Context, u domain.User, invoiceID string) (Reconciliation, error) {
	var r Reconciliation
	err := s.Store.QueryRow(ctx, "SELECT i.id,COALESCE(p.id,''),i.status,COALESCE(p.status,''),i.amount FROM invoices i LEFT JOIN payments p ON p.invoice_id=i.id WHERE i.tenant_id=? AND i.id=?", u.TenantID, invoiceID).Scan(&r.InvoiceID, &r.PaymentID, &r.InvoiceStatus, &r.PaymentStatus, &r.Amount)
	if err == sql.ErrNoRows {
		return r, domain.ErrNotFound
	}
	return r, err
}
func (s Service) RepairPayment(ctx context.Context, u domain.User, invoiceID, providerKey string) error {
	r, err := s.Reconcile(ctx, u, invoiceID)
	if err != nil {
		return err
	}
	if r.InvoiceStatus != "settled" || r.PaymentStatus == "captured" {
		return fmt.Errorf("%w: no repair needed", domain.ErrInvalidState)
	}
	_, err = s.Store.Exec(ctx, "INSERT INTO payments(id,invoice_id,provider_key,status,captured_at) VALUES(?,?,?,?,?)", storage.NewID(), invoiceID, providerKey, "captured", time.Now().UTC().Format(time.RFC3339Nano))
	return err
}
func (s Service) OpenDisputes(ctx context.Context, u domain.User) ([]string, error) {
	rows, err := s.Store.Query(ctx, "SELECT id FROM invoices WHERE tenant_id=? AND status='disputed' ORDER BY issued_at", u.TenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
