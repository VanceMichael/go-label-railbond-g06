package worker

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/outbox"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type Publisher interface {
	Publish(context.Context, string, string, string) error
}
type OutboxRunner struct {
	Store     *storage.Store
	Publisher Publisher
	Owner     string
	Epoch     int
}

func (r OutboxRunner) RunOnce(ctx context.Context, tenantID, id string) error {
	if r.Publisher == nil {
		return fmt.Errorf("publisher unavailable")
	}
	if err := (&outbox.Service{Store: r.Store}).Claim(ctx, domain.User{TenantID: tenantID}, id, r.Owner, r.Epoch, time.Now().UTC().Add(time.Minute)); err != nil {
		return err
	}
	var topic, aggregate, payload string
	if err := r.Store.QueryRow(ctx, "SELECT topic,aggregate_id,payload FROM outbox_messages WHERE tenant_id=? AND id=?", tenantID, id).Scan(&topic, &aggregate, &payload); err != nil {
		return err
	}
	if err := r.Publisher.Publish(ctx, topic, aggregate, payload); err != nil {
		if failErr := (&outbox.Service{Store: r.Store}).Fail(ctx, domain.User{TenantID: tenantID}, id, r.Owner, r.Epoch, err.Error()); failErr != nil {
			return fmt.Errorf("publish failed: %w; release lease: %v", err, failErr)
		}
		return err
	}
	return (&outbox.Service{Store: r.Store}).Acknowledge(ctx, domain.User{TenantID: tenantID}, id, r.Owner, r.Epoch)
}
func (r OutboxRunner) Drain(ctx context.Context, tenantID string, limit int) error {
	if limit < 1 {
		limit = 10
	}
	rows, err := r.Store.Query(ctx, "SELECT id FROM outbox_messages WHERE tenant_id=? AND status='pending' ORDER BY available_at LIMIT ?", tenantID, limit)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		if err := r.RunOnce(ctx, tenantID, id); err != nil {
			return err
		}
	}
	return rows.Err()
}
