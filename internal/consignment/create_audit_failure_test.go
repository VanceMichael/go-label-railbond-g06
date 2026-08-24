package consignment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/VanceMichael/go-base-railbond-g06/internal/consignment"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
)

// When the compliance audit service fails, the create request must fail and
// leave no queryable consignment or cargo items behind, so operators know they
// can safely retry the whole request.
func TestCreateAuditFailureLeavesNoBusinessRecord(t *testing.T) {
	f := testkit.New(t)
	train := f.Train(t)
	f.Store.Hooks.BeforeAudit = func(context.Context) error { return domain.ErrConflict }

	s := consignment.Service{Store: f.Store}
	_, err := s.Create(context.Background(), f.User, consignment.CreateInput{
		TrainID:     train,
		ContainerID: f.ContainerID,
		Reference:   "RB-AUDIT-FAIL",
		Items:       []consignment.Item{{SKU: "S1", Description: "tea", Quantity: 2, DeclaredValue: 10}},
	}, "req-audit-fail")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected audit failure to surface, got %v", err)
	}

	var consignments int
	if err := f.Store.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM consignments WHERE tenant_id=? AND reference=?",
		f.TenantID, "RB-AUDIT-FAIL").Scan(&consignments); err != nil || consignments != 0 {
		t.Fatalf("consignment persisted after failed audit: %v %d", err, consignments)
	}

	var items int
	if err := f.Store.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM consignment_items WHERE sku=?", "S1").Scan(&items); err != nil || items != 0 {
		t.Fatalf("items persisted after failed audit: %v %d", err, items)
	}

	var outbox int
	if err := f.Store.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM outbox_messages WHERE tenant_id=?", f.TenantID).Scan(&outbox); err != nil || outbox != 0 {
		t.Fatalf("outbox persisted after failed audit: %v %d", err, outbox)
	}

	var audits int
	if err := f.Store.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM audit_events WHERE tenant_id=? AND request_id=?",
		f.TenantID, "req-audit-fail").Scan(&audits); err != nil || audits != 0 {
		t.Fatalf("audit persisted after failed audit: %v %d", err, audits)
	}
}

// After the audit failure is cleared, the same reference can be created again.
func TestCreateRetrySucceedsAfterAuditFailure(t *testing.T) {
	f := testkit.New(t)
	train := f.Train(t)

	f.Store.Hooks.BeforeAudit = func(context.Context) error { return domain.ErrConflict }
	s := consignment.Service{Store: f.Store}
	if _, err := s.Create(context.Background(), f.User, consignment.CreateInput{
		TrainID:     train,
		ContainerID: f.ContainerID,
		Reference:   "RB-RETRY",
		Items:       []consignment.Item{{SKU: "S2", Description: "coffee", Quantity: 1}},
	}, "req-retry-1"); !errors.Is(err, domain.ErrConflict) {
		t.Fatal(err)
	}

	f.Store.Hooks.BeforeAudit = nil
	r, err := s.Create(context.Background(), f.User, consignment.CreateInput{
		TrainID:     train,
		ContainerID: f.ContainerID,
		Reference:   "RB-RETRY",
		Items:       []consignment.Item{{SKU: "S2", Description: "coffee", Quantity: 1}},
	}, "req-retry-2")
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != "booked" {
		t.Fatal(r)
	}
}
