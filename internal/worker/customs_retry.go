package worker

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/customs"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

type CustomsRetryWorker struct{ Service *customs.RetryService }

func (w CustomsRetryWorker) Process(ctx context.Context, u domain.User, id string) error {
	if err := domain.CheckContext(ctx); err != nil {
		return err
	}
	if err := w.Service.Run(ctx, u, id); err != nil {
		return fmt.Errorf("customs worker: %w", err)
	}
	return nil
}
