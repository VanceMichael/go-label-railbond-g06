package warehouse_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/warehouse"
	"testing"
)

func TestReserveAndRelease(t *testing.T) {
	f := testkit.New(t)
	id := f.Consignment(t, f.Train(t))
	s := warehouse.Service{Store: f.Store}
	if err := s.Reserve(context.Background(), f.User, f.SlotID, id, "r"); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseForCustomsRejection(context.Background(), f.User, id, "r"); err != nil {
		t.Fatal(err)
	}
	list, err := s.List(context.Background(), f.User)
	if err != nil || len(list) != 1 {
		t.Fatalf("%v %d", err, len(list))
	}
}

// When the audit write for the release fails, the slot must stay reserved and
// the reservation active so other cargo cannot grab the position. The release
// and its audit must be a single atomic transaction.
func TestReleaseRollsBackWhenAuditFails(t *testing.T) {
	f := testkit.New(t)
	id := f.Consignment(t, f.Train(t))
	s := warehouse.Service{Store: f.Store}
	if err := s.Reserve(context.Background(), f.User, f.SlotID, id, "r"); err != nil {
		t.Fatal(err)
	}

	f.Store.Hooks.BeforeAudit = func(context.Context) error { return domain.ErrConflict }
	err := s.ReleaseForCustomsRejection(context.Background(), f.User, id, "r")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict error, got %v", err)
	}

	// Slot must still be reserved for this consignment and the reservation
	// still active: the release was rolled back with the failed audit.
	if list, err := s.List(context.Background(), f.User); err != nil || len(list) != 0 {
		t.Fatalf("slot leaked after rollback: %v %d", err, len(list))
	}
	var status string
	if err := f.Store.QueryRow(context.Background(), "SELECT status FROM warehouse_slots WHERE id=?", f.SlotID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "reserved" {
		t.Fatalf("slot status = %s, want reserved", status)
	}
	exists, err := s.ReservationExists(context.Background(), f.User, id)
	if err != nil || !exists {
		t.Fatalf("reservation lost after rollback: %v %v", err, exists)
	}
}
func TestOnlyAvailableSlotsList(t *testing.T) {
	f := testkit.New(t)
	id := f.Consignment(t, f.Train(t))
	s := warehouse.Service{Store: f.Store}
	if err := s.Reserve(context.Background(), f.User, f.SlotID, id, "r"); err != nil {
		t.Fatal(err)
	}
	items, _ := s.List(context.Background(), f.User)
	if len(items) != 0 {
		t.Fatal(items)
	}
}
