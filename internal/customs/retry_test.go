package customs_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/VanceMichael/go-base-railbond-g06/internal/customs"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

// TestRetryConvergesWhenReleaseTimesOut reproduces the uncertain-remote-result
// failure: the broker has accepted the release (so Reconcile reports
// "released"), but the Release response was lost mid-flight (Release returns
// an error). The retry service must reconcile and converge the declaration to
// released instead of surfacing a failure.
func TestRetryConvergesWhenReleaseTimesOut(t *testing.T) {
	f, _, d := setup(t)
	decl := customs.Service{Store: f.Store}
	if err := decl.Submit(context.Background(), f.User, d, "r"); err != nil {
		t.Fatal(err)
	}
	broker := &customs.FakeBroker{
		ReleaseErr:      fmt.Errorf("broker response timeout"),
		ReconcileStatus: "released",
	}
	svc := &customs.RetryService{Store: f.Store, Broker: broker}
	if err := svc.Run(context.Background(), f.User, d); err != nil {
		t.Fatalf("expected convergence, got %v", err)
	}
	got, err := decl.Get(context.Background(), f.User, d)
	if err != nil || got.Status != "released" {
		t.Fatalf("status=%s err=%v", got.Status, err)
	}
}

// TestRetryReportsFailureWhenReconcileUnknown ensures that when both Release
// fails and Reconcile cannot confirm a release, the service still surfaces a
// failure (no false convergence).
func TestRetryReportsFailureWhenReconcileUnknown(t *testing.T) {
	f, _, d := setup(t)
	decl := customs.Service{Store: f.Store}
	if err := decl.Submit(context.Background(), f.User, d, "r"); err != nil {
		t.Fatal(err)
	}
	broker := &customs.FakeBroker{
		ReleaseErr:      fmt.Errorf("broker unavailable"),
		ReconcileStatus: "submitted",
	}
	svc := &customs.RetryService{Store: f.Store, Broker: broker}
	err := svc.Run(context.Background(), f.User, d)
	if err == nil {
		t.Fatal("expected failure when reconcile is not released")
	}
	if !errors.Is(err, domain.ErrInvalidState) {
		t.Fatalf("expected invalid state, got %v", err)
	}
	got, _ := decl.Get(context.Background(), f.User, d)
	if got.Status != "submitted" {
		t.Fatalf("expected submitted, got %s", got.Status)
	}
}

// TestRetryReconcilesAfterGatewayError confirms an error from Reconcile keeps
// the original broker failure surfaced (fail-safe: no spurious release).
func TestRetryReconcilesAfterGatewayError(t *testing.T) {
	f, _, d := setup(t)
	decl := customs.Service{Store: f.Store}
	if err := decl.Submit(context.Background(), f.User, d, "r"); err != nil {
		t.Fatal(err)
	}
	broker := &customs.FakeBroker{
		ReleaseErr: fmt.Errorf("broker response timeout"),
		// ReconcileStatus empty + Accept false => Reconcile returns an error.
	}
	svc := &customs.RetryService{Store: f.Store, Broker: broker}
	err := svc.Run(context.Background(), f.User, d)
	if err == nil {
		t.Fatal("expected failure when reconcile errors")
	}
	got, _ := decl.Get(context.Background(), f.User, d)
	if got.Status != "submitted" {
		t.Fatalf("expected submitted, got %s", got.Status)
	}
}

// TestRetryHappyPathAcceptsReleased keeps the direct success path covered.
func TestRetryHappyPathAcceptsReleased(t *testing.T) {
	f, _, d := setup(t)
	decl := customs.Service{Store: f.Store}
	if err := decl.Submit(context.Background(), f.User, d, "r"); err != nil {
		t.Fatal(err)
	}
	broker := &customs.FakeBroker{Accept: true}
	svc := &customs.RetryService{Store: f.Store, Broker: broker}
	if err := svc.Run(context.Background(), f.User, d); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	got, _ := decl.Get(context.Background(), f.User, d)
	if got.Status != "released" {
		t.Fatalf("status=%s", got.Status)
	}
}
