package worker_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/worker"
	"testing"
)

func TestTask0030StopPropagatesFinalizerError(t *testing.T) {
	f := testkit.New(t)
	r := &worker.Runtime{Store: f.Store}
	want := errors.New("finalizer failed")
	r.SetStopHook(func(context.Context) error { return want })
	ctx, cancel := context.WithCancel(context.Background())
	r.Start(ctx)
	cancel()
	if err := r.Stop(context.Background()); !errors.Is(err, want) {
		t.Fatalf("error=%v", err)
	}
}
