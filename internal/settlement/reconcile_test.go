package settlement_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/settlement"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestReconcileIssuedInvoice(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO invoices(id,tenant_id,consignment_id,status,amount,issued_at) VALUES(?,?,?,?,?,datetime('now'))", "r1", f.TenantID, c, "issued", 10); err != nil {
		t.Fatal(err)
	}
	r, err := (&settlement.Service{Store: f.Store}).Reconcile(context.Background(), f.Finance(), "r1")
	if err != nil || r.InvoiceStatus != "issued" {
		t.Fatalf("%v %#v", err, r)
	}
}
func TestOpenDisputes(t *testing.T) {
	f := testkit.New(t)
	ids, err := (&settlement.Service{Store: f.Store}).OpenDisputes(context.Background(), f.Finance())
	if err != nil || ids == nil {
		t.Fatalf("%v", err)
	}
}
