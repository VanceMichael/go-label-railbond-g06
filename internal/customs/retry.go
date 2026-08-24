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
	key := d.BrokerOperationKey.String
	if key == "" {
		key = storage.NewID()
		if _, err := s.Store.Exec(ctx, "UPDATE customs_declarations SET broker_operation_key=? WHERE tenant_id=? AND id=?", key, u.TenantID, id); err != nil {
			return err
		}
	}
	result, err := s.Broker.Release(ctx, id, key)
	if err != nil {
		// The broker may have accepted the release even though the response
		// was lost (e.g. a timeout mid-flight). Detach from the caller's
		// context — which may already be expired — and reconcile with the
		// broker before reporting failure so a released declaration converges.
		status, rerr := s.Broker.Reconcile(context.Background(), key)
		if rerr != nil {
			return unresolvedBrokerFailure(id, err)
		}
		if status == "released" {
			return s.markReleased(context.Background(), u, id)
		}
		return fmt.Errorf("%w: broker result %s", domain.ErrInvalidState, status)
	}
	if result == "released" {
		return s.markReleased(ctx, u, id)
	}
	return fmt.Errorf("%w: broker result %s", domain.ErrInvalidState, result)
}
func (s *RetryService) markReleased(ctx context.Context, u domain.User, id string) error {
	_, err := s.Store.Exec(ctx, "UPDATE customs_declarations SET status='released',released_at=datetime('now'),version=version+1 WHERE tenant_id=? AND id=? AND status IN ('submitted','hold')", u.TenantID, id)
	return err
}

// FakeBroker models a broker whose Release outcome and Reconcile outcome can
// diverge. ReleaseErr is non-empty to simulate a response lost mid-flight
// (e.g. a timeout after the broker accepted the release). ReconcileStatus is
// returned by Reconcile to reflect the broker's authoritative record.
type FakeBroker struct {
	Calls           int
	Mu              sync.Mutex
	Accept          bool
	ReleaseErr      error
	ReconcileStatus string
}

func (f *FakeBroker) Release(ctx context.Context, id, key string) (string, error) {
	f.Mu.Lock()
	f.Calls++
	f.Mu.Unlock()
	if f.ReleaseErr != nil {
		return "", f.ReleaseErr
	}
	if f.Accept {
		return "released", nil
	}
	return "", fmt.Errorf("broker unavailable")
}
func (f *FakeBroker) Reconcile(context.Context, string) (string, error) {
	if f.ReconcileStatus != "" {
		return f.ReconcileStatus, nil
	}
	if f.Accept {
		return "released", nil
	}
	return "", fmt.Errorf("unknown")
}

var _ = time.Second
