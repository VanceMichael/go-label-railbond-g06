package worker

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type Monitor struct{ Store *storage.Store }
type MonitorReport struct{ OutboxPending, RoutesRunning, CustomsHeld int }

func (m Monitor) Snapshot(ctx context.Context, tenantID string) (MonitorReport, error) {
	var r MonitorReport
	if err := m.Store.QueryRow(ctx, "SELECT COUNT(*) FROM outbox_messages WHERE tenant_id=? AND status='pending'", tenantID).Scan(&r.OutboxPending); err != nil {
		return r, err
	}
	if err := m.Store.QueryRow(ctx, "SELECT COUNT(*) FROM route_assignments WHERE tenant_id=? AND status='running'", tenantID).Scan(&r.RoutesRunning); err != nil {
		return r, err
	}
	if err := m.Store.QueryRow(ctx, "SELECT COUNT(*) FROM customs_declarations WHERE tenant_id=? AND status='hold'", tenantID).Scan(&r.CustomsHeld); err != nil {
		return r, err
	}
	return r, nil
}
func (m Monitor) AssertHealthy(ctx context.Context, tenantID string) error {
	r, err := m.Snapshot(ctx, tenantID)
	if err != nil {
		return err
	}
	if r.RoutesRunning > 1000 {
		return fmt.Errorf("%w: running routes", domain.ErrConflict)
	}
	return nil
}
func (m Monitor) AgeOfOldestOutbox(ctx context.Context, tenantID string) (time.Duration, error) {
	var value string
	if err := m.Store.QueryRow(ctx, "SELECT COALESCE(MIN(created_at),'') FROM outbox_messages WHERE tenant_id=? AND status='pending'", tenantID).Scan(&value); err != nil {
		return 0, err
	}
	if value == "" {
		return 0, nil
	}
	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return 0, err
	}
	return time.Since(t), nil
}
