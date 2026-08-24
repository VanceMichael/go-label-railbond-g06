package consignment_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/consignment"
	"github.com/VanceMichael/go-base-railbond-g06/internal/document"
	"github.com/VanceMichael/go-base-railbond-g06/internal/settlement"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestTask0023CancelledCloseDoesNotArchive(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c, err := (&consignment.Service{Store: f.Store}).Create(context.Background(), f.User, consignment.CreateInput{TrainID: tr, ContainerID: f.ContainerID, Reference: "close-23", Items: []consignment.Item{{SKU: "x", Description: "x", Quantity: 1, DeclaredValue: 10}}}, "r")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Store.Exec(context.Background(), "UPDATE consignments SET status='delivered' WHERE id=?", c.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO invoices(id,tenant_id,consignment_id,status,amount,issued_at) VALUES(?,?,?,?,?,datetime('now'))", "inv-23", f.TenantID, c.ID, "settled", 10); err != nil {
		t.Fatal(err)
	}
	ds := document.Service{Store: f.Store}
	if _, err := ds.Create(context.Background(), f.User, c.ID, "manifest"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Store.Exec(context.Background(), "UPDATE documents SET status='sealed' WHERE consignment_id=?", c.ID); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (&consignment.Service{Store: f.Store}).Close(ctx, f.Admin(), c.ID, "r"); err == nil {
		t.Fatal("cancelled close archived cargo")
	}
	var status string
	if err := f.Store.QueryRow(context.Background(), "SELECT status FROM consignments WHERE id=?", c.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "delivered" {
		t.Fatalf("status=%s", status)
	}
	_ = settlement.Service{}
}
