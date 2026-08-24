package dispatch_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/dispatch"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
	"time"
)

func TestContainerLeaseOwnership(t *testing.T) {
	f := testkit.New(t)
	s := dispatch.LeaseService{Store: f.Store}
	until := time.Now().UTC().Add(time.Minute)
	if err := s.ClaimContainer(context.Background(), f.User, f.ContainerID, "worker-a", "token-a", until); err != nil {
		t.Fatal(err)
	}
	if err := s.RenewContainer(context.Background(), f.User, f.ContainerID, "worker-b", "token-b", until); !errors.Is(err, domain.ErrLeaseLost) {
		t.Fatal(err)
	}
	if err := s.RenewContainer(context.Background(), f.User, f.ContainerID, "worker-a", "token-a", until.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
}
func TestRebookCreatesAssignment(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	id := f.Consignment(t, tr)
	if _, err := f.Store.Exec(context.Background(), "UPDATE consignments SET status='booked' WHERE id=?", id); err != nil {
		t.Fatal(err)
	}
	s := dispatch.RebookService{Store: f.Store}
	route, err := s.Rebook(context.Background(), f.User, id, "Carrier A", "key-1", "r")
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := f.Store.QueryRow(context.Background(), "SELECT COUNT(*) FROM route_assignments WHERE id=?", route).Scan(&n); err != nil || n != 1 {
		t.Fatalf("%v %d", err, n)
	}
}
func TestBatchDispatchRejectsHeldItem(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	id := f.Consignment(t, tr)
	if _, err := f.Store.Exec(context.Background(), "UPDATE consignments SET status='hold' WHERE id=?", id); err != nil {
		t.Fatal(err)
	}
	n, err := (&dispatch.BatchDispatcher{Store: f.Store}).Dispatch(context.Background(), f.User, tr, "r")
	if n != 0 || !errors.Is(err, domain.ErrInvalidState) {
		t.Fatalf("%d %v", n, err)
	}
}

func TestRebookReplayConvergesToFirstRoute(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	id := f.Consignment(t, tr)
	if _, err := f.Store.Exec(context.Background(), "UPDATE consignments SET status='booked' WHERE id=?", id); err != nil {
		t.Fatal(err)
	}
	s := dispatch.RebookService{Store: f.Store}
	route1, err := s.Rebook(context.Background(), f.User, id, "Carrier A", "replay-key", "r1")
	if err != nil {
		t.Fatal(err)
	}
	route2, err := s.Rebook(context.Background(), f.User, id, "Carrier A", "replay-key", "r2")
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if route1 != route2 {
		t.Fatalf("expected replay to converge onto first route %q, got %q", route1, route2)
	}
	var n int
	if err := f.Store.QueryRow(context.Background(), "SELECT COUNT(*) FROM route_assignments WHERE tenant_id=? AND consignment_id=?", f.TenantID, id).Scan(&n); err != nil || n != 1 {
		t.Fatalf("expected single assignment, got %d (%v)", n, err)
	}
}

var _ = storage.NewID
