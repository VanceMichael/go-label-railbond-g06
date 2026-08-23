package consignment_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/consignment"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
	"time"
)

func TestCloseNeedsDeliveredState(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	s := consignment.Service{Store: f.Store}
	if _, err := s.Close(context.Background(), f.Admin(), c, "r"); err == nil {
		t.Fatal("in-transit cargo closed")
	}
	if ok, err := s.CanClose(context.Background(), f.User, c); err != nil || ok {
		t.Fatalf("%v %v", err, ok)
	}
	if _, err := s.ArchiveDue(context.Background(), f.User, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
}
