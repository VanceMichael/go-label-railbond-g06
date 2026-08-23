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
		return unresolvedBrokerFailure(id, err)
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
