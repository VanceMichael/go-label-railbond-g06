package warehouse_test

import (
	"context"
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
