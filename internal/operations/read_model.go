package operations

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type Dashboard struct {
	TenantID       string
	Trains         int
	ActiveCargo    int
	HeldCustoms    int
	OpenExceptions int
	ReadyOutbox    int
	GeneratedAt    time.Time
}

type ReadModel struct {
	Store *storage.Store
}

func (r ReadModel) Dashboard(ctx context.Context, user domain.User) (Dashboard, error) {
	result := Dashboard{TenantID: user.TenantID, GeneratedAt: time.Now().UTC()}
	queries := []struct {
		target *int
		query  string
	}{
		{&result.Trains, "SELECT COUNT(*) FROM trains WHERE tenant_id=? AND status IN ('planned','published','departed')"},
		{&result.ActiveCargo, "SELECT COUNT(*) FROM consignments WHERE tenant_id=? AND status IN ('booked','in_transit','at_checkpoint')"},
		{&result.HeldCustoms, "SELECT COUNT(*) FROM customs_declarations WHERE tenant_id=? AND status='hold'"},
		{&result.OpenExceptions, "SELECT COUNT(*) FROM exceptions WHERE tenant_id=? AND status='open'"},
		{&result.ReadyOutbox, "SELECT COUNT(*) FROM outbox_messages WHERE tenant_id=? AND status='pending'"},
	}
	for _, item := range queries {
		if err := r.Store.QueryRow(ctx, item.query, user.TenantID).Scan(item.target); err != nil {
			return result, err
		}
	}
	return result, nil
}

type CargoEvent struct {
	ID          string
	Kind        string
	AggregateID string
	Status      string
	OccurredAt  time.Time
	Detail      string
}

func (r ReadModel) CargoHistory(ctx context.Context, user domain.User, id string) ([]CargoEvent, error) {
	rows, err := r.Store.Query(ctx,
		"SELECT id,action,object_id,outcome,created_at,detail FROM audit_events WHERE tenant_id=? AND object_id=? ORDER BY created_at,id",
		user.TenantID, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]CargoEvent, 0)
	for rows.Next() {
		var event CargoEvent
		var raw string
		if err := rows.Scan(&event.ID, &event.Kind, &event.AggregateID, &event.Status, &raw, &event.Detail); err != nil {
			return nil, err
		}
		parsed, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return nil, fmt.Errorf("history time: %w", err)
		}
		event.OccurredAt = parsed
		events = append(events, event)
	}
	return events, rows.Err()
}

func (r ReadModel) FindTrain(ctx context.Context, user domain.User, number string) (string, error) {
	var id string
	err := r.Store.QueryRow(ctx,
		"SELECT id FROM trains WHERE tenant_id=? AND number=?",
		user.TenantID, number,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return "", domain.ErrNotFound
	}
	return id, err
}

func (r ReadModel) CargoCountsByTrain(ctx context.Context, user domain.User) (map[string]int, error) {
	rows, err := r.Store.Query(ctx,
		"SELECT train_id,COUNT(*) FROM consignments WHERE tenant_id=? GROUP BY train_id",
		user.TenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := make(map[string]int)
	for rows.Next() {
		var trainID string
		var count int
		if err := rows.Scan(&trainID, &count); err != nil {
			return nil, err
		}
		counts[trainID] = count
	}
	return counts, rows.Err()
}

func (r ReadModel) RequireTenant(ctx context.Context, user domain.User, id string) error {
	var tenantID string
	if err := r.Store.QueryRow(ctx,
		"SELECT tenant_id FROM consignments WHERE id=?", id,
	).Scan(&tenantID); err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrNotFound
		}
		return err
	}
	if tenantID != user.TenantID {
		return domain.ErrForbidden
	}
	return nil
}

func (r ReadModel) IsReady(ctx context.Context, user domain.User) (bool, error) {
	dashboard, err := r.Dashboard(ctx, user)
	if err != nil {
		return false, err
	}
	return dashboard.HeldCustoms == 0 && dashboard.OpenExceptions < 10, nil
}

func (r ReadModel) Close(ctx context.Context) error {
	return r.Store.Ping(ctx)
}
