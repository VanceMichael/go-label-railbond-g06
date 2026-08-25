package customs

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"sync"
	"time"
)

type Broker interface {
	Release(context.Context, string, string) (string, error)
	Reconcile(context.Context, string) (string, error)
}
type RetryService struct {
	Store  *storage.Store
	Broker Broker
	mu     sync.Mutex
}

func (s *RetryService) Run(ctx context.Context, u domain.User, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := domain.CheckContext(ctx); err != nil {
		return err
	}
	d, err := s.Store.GetDeclaration(ctx, u.TenantID, id)
	if err != nil {
		return err
	}
	if d.Status != "submitted" && d.Status != "hold" {
		return fmt.Errorf("%w: declaration %s not releasable", domain.ErrInvalidState, d.Status)
	}
	// Mint exactly once and persist: every retry reuses the SAME key so the
	// external broker deduplicates the retried release as a replay.
	key, err := s.Store.AcquireBrokerOperationKey(ctx, u.TenantID, id, d.BrokerOperationKey.String)
	if err != nil {
		return err
	}
	result, err := s.Broker.Release(ctx, id, key)
	if err != nil {
		// The remote call failed (timeout, transport error). The broker may
		// still have applied the release, so reconcile by the SAME key before
		// deciding anything. Never mint a fresh key here.
		if reconcile, recErr := s.Broker.Reconcile(ctx, key); recErr == nil && reconcile == "released" {
			return s.markReleased(ctx, u, id, key)
		}
		return fmt.Errorf("broker release %s: %w", id, err)
	}
	if result == "released" {
		return s.markReleased(ctx, u, id, key)
	}
	return fmt.Errorf("%w: broker result %s", domain.ErrInvalidState, result)
}
func (s *RetryService) markReleased(ctx context.Context, u domain.User, id, key string) error {
	// Flip the declaration to released AND stamp the operation key that
	// achieved it in a single guarded update. The status precondition keeps
	// the state machine honest; pinning the key records which remote operation
	// produced the release so a later reconcile never confuses two attempts.
	_, err := s.Store.Exec(ctx, "UPDATE customs_declarations SET status='released',released_at=datetime('now'),broker_operation_key=?,version=version+1 WHERE tenant_id=? AND id=? AND status IN ('submitted','hold')", key, u.TenantID, id)
	return err
}

type FakeBroker struct {
	Calls  int
	Mu     sync.Mutex
	Accept bool
}

func (f *FakeBroker) Release(ctx context.Context, id, key string) (string, error) {
	f.Mu.Lock()
	f.Calls++
	f.Mu.Unlock()
	if f.Accept {
		return "released", nil
	}
	return "", fmt.Errorf("broker unavailable")
}
func (f *FakeBroker) Reconcile(context.Context, string) (string, error) {
	if f.Accept {
		return "released", nil
	}
	return "", fmt.Errorf("unknown")
}

var _ = time.Second
