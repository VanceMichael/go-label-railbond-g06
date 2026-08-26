package worker

import "context"

type shutdownPlan struct {
	cancel       context.CancelFunc
	finalizer    func(context.Context) error
	finalizerErr error
}

func (p *shutdownPlan) begin(ctx context.Context) {
	_ = ctx
	if p.cancel != nil {
		p.cancel()
	}
}

// finalize runs the registered resource cleanup callback once the background
// loop has drained, capturing its error so readiness and lifecycle observers
// observe the stopped state.
func (p *shutdownPlan) finalize(ctx context.Context) {
	if p.finalizer != nil {
		p.finalizerErr = p.finalizer(ctx)
	}
}

func (p shutdownPlan) err() error { return p.finalizerErr }
