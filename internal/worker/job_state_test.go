package worker_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/worker"
	"testing"
)

func TestJobTransitions(t *testing.T) {
	if err := worker.Transition(worker.JobPending, worker.JobRunning); err != nil {
		t.Fatal(err)
	}
	if err := worker.Transition(worker.JobRunning, worker.JobDead); err == nil {
		t.Fatal("invalid transition accepted")
	}
	f := testkit.New(t)
	repo := worker.JobRepository{Store: f.Store}
	if err := repo.SetRouteState(context.Background(), f.TenantID, "missing", worker.JobPending, worker.JobRunning, ""); err == nil {
		t.Fatal("missing job updated")
	}
}
