package settlement_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/consignment"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/settlement"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestSettleInvoice(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c := f.Consignment(t, tr)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO invoices(id,tenant_id,consignment_id,status,amount,issued_at) VALUES(?,?,?,?,?,datetime('now'))", "i1", f.TenantID, c, "issued", 42); err != nil {
		t.Fatal(err)
	}
	s := settlement.Service{Store: f.Store}
	if err := s.SettleInvoice(context.Background(), f.Finance(), "i1", "pay-1", "r"); err != nil {
		t.Fatal(err)
	}
	status, _, err := s.Invoice(context.Background(), f.Finance(), "i1")
	if err != nil || status != domain.InvoiceSettled {
		t.Fatalf("%v %s", err, status)
	}
}
func TestPaymentHookRollsBack(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c := f.Consignment(t, tr)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO invoices(id,tenant_id,consignment_id,status,amount,issued_at) VALUES(?,?,?,?,?,datetime('now'))", "i2", f.TenantID, c, "issued", 42); err != nil {
		t.Fatal(err)
	}
	f.Store.Hooks.BeforePayment = func(context.Context) error { return errors.New("payment down") }
	s := settlement.Service{Store: f.Store}
	if err := s.SettleInvoice(context.Background(), f.Finance(), "i2", "pay-2", "r"); err == nil {
		t.Fatal("payment error hidden")
	}
	status, _, _ := s.Invoice(context.Background(), f.Finance(), "i2")
	if status != domain.InvoiceIssued {
		t.Fatal(status)
	}
}
func TestDisputeNeedsFinanceRole(t *testing.T) {
	f := testkit.New(t)
	s := settlement.Service{Store: f.Store}
	if err := s.Dispute(context.Background(), f.User, "i", "x"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal(err)
	}
}

var _ = consignment.Item{}
