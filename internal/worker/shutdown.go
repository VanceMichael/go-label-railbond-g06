package worker

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

func PersistCancellation(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return domain.ErrCancelled
	}
	return err
}
