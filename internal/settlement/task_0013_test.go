package settlement_test

import (
	"context"
	"errors"
	"testing"

	"github.com/VanceMichael/go-base-railbond-g06/internal/consignment"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
)

func TestTask0013DeliveryFailureRollsBackInvoiceAndStatus(t *testing.T) {
	f := testkit.New(t)
	trainID := f.Train(t)
	failedConsignmentID := f.Consignment(t, trainID)
	f.Store.Hooks.BeforeOutbox = func(context.Context) error { return errors.New("outbox down") }

	if _, err := (&consignment.DeliveryService{Store: f.Store}).Deliver(context.Background(), f.User, failedConsignmentID, "receiver", "request-failed"); err == nil {
		t.Fatal("delivery succeeded while its durable event failed")
	}
	failedRecord, err := (&consignment.Service{Store: f.Store}).Get(context.Background(), f.User, failedConsignmentID)
	if err != nil {
		t.Fatal(err)
	}
	if failedRecord.Status != "in_transit" {
		t.Fatalf("failed delivery left consignment status=%s", failedRecord.Status)
	}
	var invoiceCount int
	if err := f.Store.QueryRow(context.Background(), "SELECT COUNT(*) FROM invoices WHERE consignment_id=?", failedConsignmentID).Scan(&invoiceCount); err != nil {
		t.Fatal(err)
	}
	if invoiceCount != 0 {
		t.Fatalf("failed delivery left %d invoice rows", invoiceCount)
	}

	f.Store.Hooks.BeforeOutbox = nil
	successfulConsignmentID := f.Consignment(t, trainID)
	invoiceID, err := (&consignment.DeliveryService{Store: f.Store}).Deliver(context.Background(), f.User, successfulConsignmentID, "receiver", "request-success")
	if err != nil {
		t.Fatalf("normal delivery failed: %v", err)
	}
	successfulRecord, err := (&consignment.Service{Store: f.Store}).Get(context.Background(), f.User, successfulConsignmentID)
	if err != nil {
		t.Fatal(err)
	}
	if successfulRecord.Status != "delivered" || invoiceID == "" {
		t.Fatalf("normal delivery status=%s invoice=%q", successfulRecord.Status, invoiceID)
	}
}
