package warehouse_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/warehouse"
	"testing"
)

func TestMoveReservation(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	s := warehouse.Service{Store: f.Store}
	if err := s.Reserve(context.Background(), f.User, f.SlotID, c, "r"); err != nil {
		t.Fatal(err)
	}
	second := "slot-2"
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO warehouse_slots(id,tenant_id,code,zone,status,version) VALUES(?,?,?,?,?,1)", second, f.TenantID, "A-02", "bonded", "available"); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(context.Background(), f.User, f.SlotID, second, c, "r"); err != nil {
		t.Fatal(err)
	}
	slot, err := s.Reservation(context.Background(), f.User, c)
	if err != nil || slot != second {
		t.Fatalf("%v %s", err, slot)
	}
}

func TestMoveAtomicOnAuditFailure(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	s := warehouse.Service{Store: f.Store}
	if err := s.Reserve(context.Background(), f.User, f.SlotID, c, "r"); err != nil {
		t.Fatal(err)
	}
	second := "slot-2"
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO warehouse_slots(id,tenant_id,code,zone,status,version) VALUES(?,?,?,?,?,1)", second, f.TenantID, "A-02", "bonded", "available"); err != nil {
		t.Fatal(err)
	}
	auditErr := errors.New("audit unavailable")
	f.Store.Hooks.BeforeAudit = func(context.Context) error { return auditErr }
	defer func() { f.Store.Hooks.BeforeAudit = nil }()

	if err := s.Move(context.Background(), f.User, f.SlotID, second, c, "r"); !errors.Is(err, auditErr) {
		t.Fatalf("expected audit error, got %v", err)
	}
	if slot, _ := s.Reservation(context.Background(), f.User, c); slot != f.SlotID {
		t.Fatalf("reservation should point to original slot, got %s", slot)
	}
	var fromStatus, toStatus string
	if err := f.Store.QueryRow(context.Background(), "SELECT status FROM warehouse_slots WHERE id=?", f.SlotID).Scan(&fromStatus); err != nil {
		t.Fatal(err)
	}
	if err := f.Store.QueryRow(context.Background(), "SELECT status FROM warehouse_slots WHERE id=?", second).Scan(&toStatus); err != nil {
		t.Fatal(err)
	}
	if fromStatus != "reserved" {
		t.Fatalf("source slot should remain reserved, got %s", fromStatus)
	}
	if toStatus != "available" {
		t.Fatalf("target slot should remain available, got %s", toStatus)
	}
}
func TestOccupancy(t *testing.T) {
	f := testkit.New(t)
	o, err := (&warehouse.Service{Store: f.Store}).Occupancy(context.Background(), f.User, "bonded")
	if err != nil || o.Available != 1 {
		t.Fatalf("%v %#v", err, o)
	}
}
