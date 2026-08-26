package dispatch_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/dispatch"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
	"time"
)

func TestTask0005StaleWorkerCannotReleaseCurrentContainerLease(t *testing.T) {
	f := testkit.New(t)
	s := dispatch.LeaseService{Store: f.Store}
	if err := s.ClaimContainer(context.Background(), f.User, f.ContainerID, "worker-b", "token-b", time.Now().UTC().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseContainer(context.Background(), f.User, f.ContainerID, "worker-a", "token-a"); err == nil {
		t.Fatal("stale worker released current lease")
	}
	var owner, token string
	_ = f.Store.QueryRow(context.Background(), "SELECT lease_owner,lease_token FROM containers WHERE id=?", f.ContainerID).Scan(&owner, &token)
	if owner != "worker-b" || token != "token-b" {
		t.Fatalf("lease changed owner=%s token=%s", owner, token)
	}
}
