package lifecycle

import (
	"context"
	"fmt"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type Preflight struct {
	Store *storage.Store
}

type Result struct {
	Name   string
	Passed bool
	Detail string
}

func (p Preflight) Run(ctx context.Context, user domain.User, trainID string) ([]Result, error) {
	results := make([]Result, 0, 5)
	var status string
	var departure string
	var capacity, reserved int
	if err := p.Store.QueryRow(ctx,
		"SELECT status,departure_at,capacity,reserved FROM trains WHERE tenant_id=? AND id=?",
		user.TenantID, trainID,
	).Scan(&status, &departure, &capacity, &reserved); err != nil {
		return nil, err
	}
	results = append(results, Result{
		Name:   "train_status",
		Passed: status == "published",
		Detail: status,
	})
	when, err := time.Parse(time.RFC3339Nano, departure)
	if err != nil {
		return nil, err
	}
	results = append(results, Result{
		Name:   "departure_window",
		Passed: when.After(time.Now().UTC()),
		Detail: when.Format(time.RFC3339Nano),
	})
	results = append(results, Result{
		Name:   "capacity",
		Passed: capacity >= reserved,
		Detail: fmt.Sprintf("%d/%d", reserved, capacity),
	})
	var customs int
	if err := p.Store.QueryRow(ctx,
		"SELECT COUNT(*) FROM customs_declarations d JOIN consignments c ON c.id=d.consignment_id WHERE c.tenant_id=? AND c.train_id=? AND d.status!='released'",
		user.TenantID, trainID,
	).Scan(&customs); err != nil {
		return nil, err
	}
	results = append(results, Result{
		Name:   "customs",
		Passed: customs == 0,
		Detail: fmt.Sprintf("pending=%d", customs),
	})
	var exceptions int
	if err := p.Store.QueryRow(ctx,
		"SELECT COUNT(*) FROM exceptions e JOIN consignments c ON c.id=e.consignment_id WHERE e.tenant_id=? AND c.train_id=? AND e.status='open'",
		user.TenantID, trainID,
	).Scan(&exceptions); err != nil {
		return nil, err
	}
	results = append(results, Result{
		Name:   "exceptions",
		Passed: exceptions == 0,
		Detail: fmt.Sprintf("open=%d", exceptions),
	})
	return results, nil
}

func (p Preflight) Passed(results []Result) bool {
	if len(results) == 0 {
		return false
	}
	for _, result := range results {
		if !result.Passed {
			return false
		}
	}
	return true
}

func (p Preflight) Record(ctx context.Context, user domain.User, trainID, requestID string, results []Result) error {
	if !p.Passed(results) {
		return fmt.Errorf("%w: preflight failed", domain.ErrInvalidState)
	}
	return p.Store.WithTx(ctx, func(tx *storage.Tx) error {
		return p.Store.RecordAudit(ctx, tx, user.TenantID, user.ID,
			"train.preflight", "train", trainID, "success", requestID,
			fmt.Sprintf("checks=%d", len(results)))
	})
}

func (p Preflight) RunAndRecord(ctx context.Context, user domain.User, trainID, requestID string) (bool, error) {
	results, err := p.Run(ctx, user, trainID)
	if err != nil {
		return false, err
	}
	if err := p.Record(ctx, user, trainID, requestID, results); err != nil {
		return false, err
	}
	return true, nil
}
