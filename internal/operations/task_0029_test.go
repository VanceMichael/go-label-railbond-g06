package operations_test

import (
	"context"
	"testing"

	"github.com/VanceMichael/go-base-railbond-g06/internal/operations"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
)

func TestTask0029CommitDispatchRollsBackOnLateConflict(t *testing.T) {
	f := testkit.New(t)
	trainID := f.Train(t)
	firstID := f.Consignment(t, trainID)
	secondID := f.Consignment(t, trainID)
	for _, id := range []string{firstID, secondID} {
		if _, err := f.Store.Exec(context.Background(), "UPDATE consignments SET status='booked' WHERE id=?", id); err != nil {
			t.Fatal(err)
		}
	}
	plan := operations.DispatchPlan{TrainID: trainID, Consignments: []string{firstID, secondID}, Ready: true}
	if _, err := f.Store.Exec(context.Background(), "UPDATE consignments SET status='draft' WHERE id=?", secondID); err != nil {
		t.Fatal(err)
	}
	coordinator := operations.Coordinator{Store: f.Store}
	if err := coordinator.CommitDispatch(context.Background(), f.User, plan, "conflicted-dispatch"); err == nil {
		t.Fatal("late conflict was accepted")
	}
	assertDispatchState(t, f, firstID, secondID, "booked", "draft", 0)

	if _, err := f.Store.Exec(context.Background(), "UPDATE consignments SET status='booked' WHERE id=?", secondID); err != nil {
		t.Fatal(err)
	}
	if err := coordinator.CommitDispatch(context.Background(), f.User, plan, "successful-dispatch"); err != nil {
		t.Fatalf("valid dispatch failed: %v", err)
	}
	assertDispatchState(t, f, firstID, secondID, "in_transit", "in_transit", 1)
}

func assertDispatchState(t *testing.T, f testkit.Fixture, firstID, secondID, firstStatus, secondStatus string, auditCount int) {
	t.Helper()
	var first, second string
	if err := f.Store.QueryRow(context.Background(), "SELECT status FROM consignments WHERE id=?", firstID).Scan(&first); err != nil {
		t.Fatal(err)
	}
	if err := f.Store.QueryRow(context.Background(), "SELECT status FROM consignments WHERE id=?", secondID).Scan(&second); err != nil {
		t.Fatal(err)
	}
	var audits int
	if err := f.Store.QueryRow(context.Background(), "SELECT COUNT(*) FROM audit_events WHERE action='train.dispatched'").Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if first != firstStatus || second != secondStatus || audits != auditCount {
		t.Fatalf("dispatch state first=%s second=%s audits=%d", first, second, audits)
	}
}
