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

func (p shutdownPlan) err() error { return p.finalizerErr }
