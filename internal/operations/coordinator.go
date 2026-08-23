package operations

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type Coordinator struct {
	Store *storage.Store
}

type Check struct {
	Name   string
	Passed bool
	Detail string
}

type DispatchPlan struct {
	TrainID      string
	Consignments []string
	Checks       []Check
	CreatedAt    time.Time
	Ready        bool
}

func (c Coordinator) BuildDispatchPlan(ctx context.Context, user domain.User, trainID string) (DispatchPlan, error) {
	plan := DispatchPlan{TrainID: trainID, CreatedAt: time.Now().UTC()}
	rows, err := c.Store.Query(ctx,
		"SELECT id FROM consignments WHERE tenant_id=? AND train_id=? AND status='booked' ORDER BY id",
		user.TenantID, trainID,
	)
	if err != nil {
		return plan, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return plan, err
		}
		plan.Consignments = append(plan.Consignments, id)
	}
	if err := rows.Err(); err != nil {
		return plan, err
	}
	trainCheck, err := c.checkTrain(ctx, user, trainID)
	if err != nil {
		return plan, err
	}
	plan.Checks = append(plan.Checks, trainCheck)
	for _, id := range plan.Consignments {
		checks, err := c.checkConsignment(ctx, user, id)
		if err != nil {
			return plan, err
		}
		plan.Checks = append(plan.Checks, checks...)
	}
	plan.Ready = len(plan.Consignments) > 0
	for _, check := range plan.Checks {
		if !check.Passed {
			plan.Ready = false
		}
	}
	return plan, nil
}

func (c Coordinator) checkTrain(ctx context.Context, user domain.User, trainID string) (Check, error) {
	var status string
	var capacity, reserved int
	err := c.Store.QueryRow(ctx,
		"SELECT status,capacity,reserved FROM trains WHERE tenant_id=? AND id=?",
		user.TenantID, trainID,
	).Scan(&status, &capacity, &reserved)
	if err != nil {
		if err == sql.ErrNoRows {
			return Check{}, domain.ErrNotFound
		}
		return Check{}, err
	}
	passed := status == "published" && capacity >= reserved
	detail := fmt.Sprintf("status=%s reserved=%d capacity=%d", status, reserved, capacity)
	return Check{Name: "train_capacity", Passed: passed, Detail: detail}, nil
}

func (c Coordinator) checkConsignment(ctx context.Context, user domain.User, id string) ([]Check, error) {
	var customs string
	if err := c.Store.QueryRow(ctx,
		"SELECT COALESCE((SELECT status FROM customs_declarations WHERE tenant_id=? AND consignment_id=? LIMIT 1),'')",
		user.TenantID, id,
	).Scan(&customs); err != nil {
		return nil, err
	}
	var items int
	if err := c.Store.QueryRow(ctx,
		"SELECT COUNT(*) FROM consignment_items WHERE consignment_id=?",
		id,
	).Scan(&items); err != nil {
		return nil, err
	}
	var holds int
	if err := c.Store.QueryRow(ctx,
		"SELECT COUNT(*) FROM exceptions WHERE tenant_id=? AND consignment_id=? AND status='open'",
		user.TenantID, id,
	).Scan(&holds); err != nil {
		return nil, err
	}
	return []Check{
		{Name: "customs_release", Passed: customs == "released", Detail: customs},
		{Name: "declared_items", Passed: items > 0, Detail: fmt.Sprintf("items=%d", items)},
		{Name: "open_exception", Passed: holds == 0, Detail: fmt.Sprintf("open=%d", holds)},
	}, nil
}

func (c Coordinator) CommitDispatch(ctx context.Context, user domain.User, plan DispatchPlan, requestID string) error {
	if !plan.Ready {
		return fmt.Errorf("%w: dispatch plan is not ready", domain.ErrInvalidState)
	}
	return c.Store.WithTx(ctx, func(tx *storage.Tx) error {
		for _, id := range plan.Consignments {
			if _, err := tx.Exec(ctx,
				"UPDATE consignments SET status='in_transit',version=version+1 WHERE tenant_id=? AND id=? AND status='booked'",
				user.TenantID, id,
			); err != nil {
				return err
			}
		}
		detail := fmt.Sprintf("consignments=%d", len(plan.Consignments))
		return c.Store.RecordAudit(ctx, tx, user.TenantID, user.ID,
			"train.dispatched", "train", plan.TrainID, "success", requestID, detail)
	})
}

func (c Coordinator) CancelPlan(ctx context.Context, user domain.User, plan DispatchPlan, requestID string) error {
	if len(plan.Consignments) == 0 {
		return nil
	}
	return c.Store.WithTx(ctx, func(tx *storage.Tx) error {
		for _, id := range plan.Consignments {
			if _, err := tx.Exec(ctx,
				"UPDATE consignments SET status='draft',version=version+1 WHERE tenant_id=? AND id=? AND status='booked'",
				user.TenantID, id,
			); err != nil {
				return err
			}
		}
		return c.Store.RecordAudit(ctx, tx, user.TenantID, user.ID,
			"train.dispatch_cancelled", "train", plan.TrainID, "success", requestID,
			fmt.Sprintf("consignments=%d", len(plan.Consignments)))
	})
}

func (c Coordinator) ValidatePlan(plan DispatchPlan) error {
	if plan.TrainID == "" {
		return fmt.Errorf("%w: missing train", domain.ErrInvalidState)
	}
	if len(plan.Consignments) == 0 {
		return fmt.Errorf("%w: empty plan", domain.ErrInvalidState)
	}
	seen := make(map[string]struct{}, len(plan.Consignments))
	for _, id := range plan.Consignments {
		if id == "" {
			return fmt.Errorf("%w: empty consignment", domain.ErrInvalidState)
		}
		if _, ok := seen[id]; ok {
			return fmt.Errorf("%w: duplicate consignment", domain.ErrConflict)
		}
		seen[id] = struct{}{}
	}
	return nil
}

type RecoveryItem struct {
	ID          string
	OldStatus   string
	NextStatus  string
	Reason      string
	Recoverable bool
}

func (c Coordinator) RecoveryQueue(ctx context.Context, user domain.User) ([]RecoveryItem, error) {
	rows, err := c.Store.Query(ctx,
		"SELECT id,status,last_error FROM route_assignments WHERE tenant_id=? AND status IN ('retry','dead') ORDER BY next_attempt_at,id",
		user.TenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]RecoveryItem, 0)
	for rows.Next() {
		var item RecoveryItem
		if err := rows.Scan(&item.ID, &item.OldStatus, &item.Reason); err != nil {
			return nil, err
		}
		item.NextStatus = "assigned"
		item.Recoverable = item.OldStatus == "retry"
		items = append(items, item)
	}
	return items, rows.Err()
}

func (c Coordinator) Requeue(ctx context.Context, user domain.User, id, requestID string) error {
	return c.Store.WithTx(ctx, func(tx *storage.Tx) error {
		res, err := tx.Exec(ctx,
			"UPDATE route_assignments SET status='assigned',attempt=attempt+1,next_attempt_at=?,last_error=NULL WHERE tenant_id=? AND id=? AND status='retry'",
			time.Now().UTC().Format(time.RFC3339Nano), user.TenantID, id,
		)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n != 1 {
			return fmt.Errorf("%w: route is not retryable", domain.ErrInvalidState)
		}
		return c.Store.RecordAudit(ctx, tx, user.TenantID, user.ID,
			"route.requeued", "route_assignment", id, "success", requestID, "retry")
	})
}
