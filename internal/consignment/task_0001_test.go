package consignment_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/consignment"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestTask0001AuditFailureRollsBackConsignment(t *testing.T) {
	f := testkit.New(t)
	f.Store.Hooks.BeforeAudit = func(context.Context) error { return errors.New("audit unavailable") }
	s := consignment.Service{Store: f.Store}
	_, err := s.Create(context.Background(), f.User, consignment.CreateInput{TrainID: f.Train(t), ContainerID: f.ContainerID, Reference: "rollback-audit", Items: []consignment.Item{{SKU: "coffee", Description: "beans", Quantity: 1, DeclaredValue: 10}}}, "task-0001")
	if err == nil {
		t.Fatal("expected audit failure")
	}
	var n int
	if err := f.Store.QueryRow(context.Background(), "SELECT COUNT(*) FROM consignments WHERE tenant_id=? AND reference=?", f.TenantID, "rollback-audit").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("failed creation left %d consignment rows", n)
	}
}
