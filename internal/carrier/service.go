package carrier

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"strings"
	"sync"
	"time"
)

type Receipt struct {
	ProviderKey, Status, TrackingURL string
	ReceivedAt                       time.Time
}
type Client interface {
	Accept(context.Context, string, string) (Receipt, error)
	Cancel(context.Context, string, string) error
}
type Service struct {
	Store  *storage.Store
	Client Client
	mu     sync.Mutex
}

func (s *Service) Accept(ctx context.Context, u domain.User, assignmentID, providerKey string) (Receipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := domain.CheckContext(ctx); err != nil {
		return Receipt{}, err
	}
	if strings.TrimSpace(providerKey) == "" {
		return Receipt{}, fmt.Errorf("%w: provider key", domain.ErrInvalidState)
	}
	var status, tenant string
	if err := s.Store.QueryRow(ctx, "SELECT status,tenant_id FROM route_assignments WHERE id=?", assignmentID).Scan(&status, &tenant); err != nil {
		return Receipt{}, err
	}
	if tenant != u.TenantID {
		return Receipt{}, domain.ErrForbidden
	}
	if status != "assigned" && status != "retry" {
		return Receipt{}, fmt.Errorf("%w: assignment status", domain.ErrInvalidState)
	}
	if s.Client == nil {
		return Receipt{}, fmt.Errorf("carrier client unavailable")
	}
	receipt, err := s.Client.Accept(ctx, assignmentID, providerKey)
	if err != nil {
		return Receipt{}, fmt.Errorf("carrier accept: %w", err)
	}
	if receipt.Status != "accepted" {
		return Receipt{}, fmt.Errorf("%w: carrier result", domain.ErrInvalidState)
	}
	if _, err := s.Store.Exec(ctx, "UPDATE route_assignments SET status='running',lease_owner=?,lease_until=? WHERE id=? AND status IN ('assigned','retry')", u.ID, time.Now().UTC().Add(time.Minute).Format(time.RFC3339Nano), assignmentID); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}
func (s *Service) Cancel(ctx context.Context, u domain.User, assignmentID, providerKey string) error {
	if err := domain.CheckContext(ctx); err != nil {
		return err
	}
	if s.Client == nil {
		return fmt.Errorf("carrier client unavailable")
	}
	if err := s.Client.Cancel(ctx, assignmentID, providerKey); err != nil {
		return fmt.Errorf("carrier cancel: %w", err)
	}
	res, err := s.Store.Exec(ctx, "UPDATE route_assignments SET status='cancelled',lease_owner=NULL,lease_until=NULL WHERE tenant_id=? AND id=? AND status IN ('assigned','running')", u.TenantID, assignmentID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}
