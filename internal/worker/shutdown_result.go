package worker

type shutdownOutcome struct {
	finalizerErr error
	contextErr   error
}

func (o shutdownOutcome) Err() error {
	if o.contextErr != nil {
		return o.contextErr
	}
	return nil
}
