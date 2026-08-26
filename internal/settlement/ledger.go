package settlement

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type LedgerEntry struct {
	ID, InvoiceID, Kind string
	Amount              int
}

func (s Service) Ledger(ctx context.Context, u domain.User, invoiceID string) ([]LedgerEntry, error) {
	rows, err := s.Store.Query(ctx, "SELECT id,invoice_id,status,amount FROM payments WHERE invoice_id=? ORDER BY id", invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []LedgerEntry{}
	for rows.Next() {
		var e LedgerEntry
		if err := rows.Scan(&e.ID, &e.InvoiceID, &e.Kind, &e.Amount); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
func (s Service) AssertBalanced(ctx context.Context, u domain.User, invoiceID string) error {
	status, amount, err := s.Invoice(ctx, u, invoiceID)
	if err != nil {
		return err
	}
	entries, err := s.Ledger(ctx, u, invoiceID)
	if err != nil {
		return err
	}
	if status == domain.InvoiceSettled && len(entries) == 0 {
		return fmt.Errorf("%w: settled invoice has no payment", domain.ErrConflict)
	}
	for _, e := range entries {
		if e.Amount != amount && amount != 0 {
			return fmt.Errorf("%w: payment amount", domain.ErrConflict)
		}
	}
	return nil
}
func (s Service) AddAdjustment(ctx context.Context, u domain.User, invoiceID string, amount int) error {
	if amount == 0 {
		return domain.ErrInvalidState
	}
	_, err := s.Store.Exec(ctx, "INSERT INTO payments(id,invoice_id,provider_key,status,amount) VALUES(?,?,?,?,?)", storage.NewID(), invoiceID, "adjustment-"+storage.NewID(), "adjustment", amount)
	return err
}
