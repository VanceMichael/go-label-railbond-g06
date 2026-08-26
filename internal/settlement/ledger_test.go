package settlement_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/settlement"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestLedgerEmptyForIssued(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO invoices(id,tenant_id,consignment_id,status,amount,issued_at) VALUES(?,?,?,?,?,datetime('now'))", "l1", f.TenantID, c, "issued", 10); err != nil {
		t.Fatal(err)
	}
	items, err := (&settlement.Service{Store: f.Store}).Ledger(context.Background(), f.Finance(), "l1")
	if err != nil || len(items) != 0 {
		t.Fatalf("%v %#v", err, items)
	}
}
