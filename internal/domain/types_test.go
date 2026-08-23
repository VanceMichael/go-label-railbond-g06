package domain_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"testing"
)

func TestStateMachinesRejectIllegalMoves(t *testing.T) {
	if domain.TrainPlanned.CanMove(domain.TrainArrived) {
		t.Fatal("train skipped states")
	}
	if domain.ConsignmentDraft.CanMove(domain.ConsignmentDelivered) {
		t.Fatal("consignment skipped states")
	}
	if domain.DeclarationDraft.CanMove(domain.DeclarationReleased) {
		t.Fatal("declaration skipped states")
	}
}
func TestErrorWrappingAndCancellation(t *testing.T) {
	wrapped := domain.Wrap("operation", domain.ErrConflict)
	if !errors.Is(wrapped, domain.ErrConflict) {
		t.Fatal(wrapped)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !errors.Is(domain.CheckContext(ctx), domain.ErrCancelled) {
		t.Fatal("cancel not preserved")
	}
}
func TestRoles(t *testing.T) {
	for _, role := range []string{"operator", "customs", "finance", "admin"} {
		if err := domain.RequireRole(domain.User{Role: role}, role); err != nil {
			t.Fatal(role, err)
		}
	}
}
