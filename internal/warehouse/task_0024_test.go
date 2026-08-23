package warehouse_test

import (
	"context"
	"errors"
	"testing"

	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/warehouse"
)

func TestTask0024MoveAuditFailureRollsBackBothSlots(t *testing.T) {
	f := testkit.New(t)
	consignmentID := f.Consignment(t, f.Train(t))
	service := warehouse.Service{Store: f.Store}
	if err := service.Reserve(context.Background(), f.User, f.SlotID, consignmentID, "reserve-request"); err != nil {
		t.Fatal(err)
	}
	destinationID := "slot-24"
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO warehouse_slots(id,tenant_id,code,zone,status,version) VALUES(?,?,?,?,?,1)", destinationID, f.TenantID, "B-24", "bonded", "available"); err != nil {
		t.Fatal(err)
	}
	f.Store.Hooks.BeforeAudit = func(context.Context) error { return errors.New("audit down") }
	if err := service.Move(context.Background(), f.User, f.SlotID, destinationID, consignmentID, "failed-move"); err == nil {
		t.Fatal("move succeeded while its audit failed")
	}
	assertMoveState(t, f, service, consignmentID, f.SlotID, destinationID, "reserved", "available", f.SlotID)

	f.Store.Hooks.BeforeAudit = nil
	if err := service.Move(context.Background(), f.User, f.SlotID, destinationID, consignmentID, "successful-move"); err != nil {
		t.Fatalf("move failed after audit recovery: %v", err)
	}
	assertMoveState(t, f, service, consignmentID, f.SlotID, destinationID, "available", "reserved", destinationID)
}

func assertMoveState(t *testing.T, f testkit.Fixture, service warehouse.Service, consignmentID, sourceID, destinationID, sourceStatus, destinationStatus, reservationID string) {
	t.Helper()
	var source, destination string
	if err := f.Store.QueryRow(context.Background(), "SELECT status FROM warehouse_slots WHERE id=?", sourceID).Scan(&source); err != nil {
		t.Fatal(err)
	}
	if err := f.Store.QueryRow(context.Background(), "SELECT status FROM warehouse_slots WHERE id=?", destinationID).Scan(&destination); err != nil {
		t.Fatal(err)
	}
	reservedAt, err := service.Reservation(context.Background(), f.User, consignmentID)
	if err != nil {
		t.Fatal(err)
	}
	if source != sourceStatus || destination != destinationStatus || reservedAt != reservationID {
		t.Fatalf("move state source=%s destination=%s reservation=%s", source, destination, reservedAt)
	}
}
