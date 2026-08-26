package worker_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/worker"
	"testing"
)

func TestTask0015StopRunsFinalizationHook(t *testing.T) {
	f := testkit.New(t)
	r := &worker.Runtime{Store: f.Store}
	called := false
	r.SetStopHook(func(context.Context) error { called = true; return nil })
	ctx, cancel := context.WithCancel(context.Background())
	r.Start(ctx)
	cancel()
	if err := r.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("stop finalization hook was skipped")
	}
}
