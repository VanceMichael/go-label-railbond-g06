package warehouse_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/warehouse"
	"testing"
)

func TestTask0008CustomsRejectionRollbackKeepsSlot(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	s := warehouse.Service{Store: f.Store}
	if err := s.Reserve(context.Background(), f.User, f.SlotID, c, "r"); err != nil {
		t.Fatal(err)
	}
	f.Store.Hooks.BeforeAudit = func(context.Context) error { return errors.New("audit down") }
	if err := s.ReleaseForCustomsRejection(context.Background(), f.User, c, "r"); err == nil {
		t.Fatal("release succeeded")
	}
	var status, reserved string
	if err := f.Store.QueryRow(context.Background(), "SELECT ws.status,COALESCE(ws.reserved_for,'') FROM warehouse_slots ws WHERE ws.id=?", f.SlotID).Scan(&status, &reserved); err != nil {
		t.Fatal(err)
	}
	if status != "reserved" || reserved != c {
		t.Fatalf("slot changed status=%s reserved_for=%s", status, reserved)
	}
	var rs string
	if err := f.Store.QueryRow(context.Background(), "SELECT status FROM slot_reservations WHERE consignment_id=?", c).Scan(&rs); err != nil {
		t.Fatal(err)
	}
	if rs != "active" {
		t.Fatalf("reservation=%s", rs)
	}
}
