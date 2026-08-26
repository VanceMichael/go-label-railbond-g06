package worker

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"sync"
	"time"
)

type Runtime struct {
	Store  *storage.Store
	cancel context.CancelFunc
	wg     sync.WaitGroup
	once   sync.Once
	onStop func(context.Context) error
}

func (r *Runtime) Start(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	r.cancel = cancel
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.recoverExpired(ctx)
			}
		}
	}()
}
func (r *Runtime) recoverExpired(ctx context.Context) {
	_, _ = r.Store.Exec(ctx, "UPDATE outbox_messages SET status='pending',lease_owner=NULL,lease_until=NULL WHERE status='sending' AND lease_until<?", time.Now().UTC().Format(time.RFC3339Nano))
	_, _ = r.Store.Exec(ctx, "UPDATE route_assignments SET status='retry',lease_owner=NULL,lease_until=NULL WHERE status='running' AND lease_until<?", time.Now().UTC().Format(time.RFC3339Nano))
}
func (r *Runtime) Stop(ctx context.Context) error {
	var plan shutdownPlan
	r.once.Do(func() {
		plan = shutdownPlan{cancel: r.cancel, finalizer: r.onStop}
		plan.begin(ctx)
	})
	done := make(chan struct{})
	go func() { r.wg.Wait(); close(done) }()
	select {
	case <-done:
		plan.finalize(ctx)
		return plan.err()
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (r *Runtime) SetStopHook(fn func(context.Context) error) { r.onStop = fn }
