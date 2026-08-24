package worker

import "errors"

type shutdownOutcome struct {
	finalizerErr error
	contextErr   error
}

func (o shutdownOutcome) Err() error {
	var errs []error
	if o.finalizerErr != nil {
		errs = append(errs, o.finalizerErr)
	}
	if o.contextErr != nil {
		errs = append(errs, o.contextErr)
	}
	return errors.Join(errs...)
}
