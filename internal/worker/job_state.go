package worker

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type JobState string

const (
	JobPending JobState = "pending"
	JobRunning JobState = "running"
	JobRetry   JobState = "retry"
	JobDead    JobState = "dead"
)

func Transition(from, to JobState) error {
	if from == JobPending && to == JobRunning || from == JobRunning && to == JobRetry || from == JobRetry && to == JobRunning || from == JobRetry && to == JobDead {
		return nil
	}
	return fmt.Errorf("%w: worker state %s to %s", domain.ErrInvalidState, from, to)
}
func ClaimUntil(now time.Time, lease time.Duration) time.Time {
	if lease <= 0 {
		lease = time.Minute
	}
	return now.Add(lease)
}

type JobRepository struct{ Store *storage.Store }

func (r JobRepository) SetRouteState(ctx context.Context, tenantID, id string, from, to JobState, reason string) error {
	if err := Transition(from, to); err != nil {
		return err
	}
	res, err := r.Store.Exec(ctx, "UPDATE route_assignments SET status=?,last_error=? WHERE tenant_id=? AND id=? AND status=?", to, reason, tenantID, id, from)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}
