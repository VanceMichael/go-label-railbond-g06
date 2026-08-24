package dispatch_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/dispatch"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestTask0003RebookReplayDoesNotDuplicateRoute(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c := f.Consignment(t, tr)
	if _, err := f.Store.Exec(context.Background(), "UPDATE consignments SET status='booked' WHERE id=?", c); err != nil {
		t.Fatal(err)
	}
	s := dispatch.RebookService{Store: f.Store}
	first, err := s.Rebook(context.Background(), f.User, c, "Carrier", "rebook-key", "r")
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.Rebook(context.Background(), f.User, c, "Carrier", "rebook-key", "r")
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("replay assigned %s after %s", second, first)
	}
	var n int
	_ = f.Store.QueryRow(context.Background(), "SELECT COUNT(*) FROM route_assignments WHERE consignment_id=?", c).Scan(&n)
	if n != 1 {
		t.Fatalf("route assignments=%d", n)
	}
}
