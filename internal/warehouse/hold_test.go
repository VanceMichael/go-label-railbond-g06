package warehouse_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/warehouse"
	"testing"
)

func TestSlotHoldLifecycle(t *testing.T) {
	f := testkit.New(t)
	s := warehouse.Service{Store: f.Store}
	if err := s.PlaceHold(context.Background(), f.User, f.SlotID, "inspection"); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseHold(context.Background(), f.User, f.SlotID); err != nil {
		t.Fatal(err)
	}
	c := f.Consignment(t, f.Train(t))
	if err := s.HoldForConsignment(context.Background(), f.User, f.SlotID, c, "customs"); err != nil {
		t.Fatal(err)
	}
}
