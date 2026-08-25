package customs_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/VanceMichael/go-base-railbond-g06/internal/customs"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

// recordingBroker captures every key it receives across retries so the test
// can assert that a retried release reuses the identical operation key.
type recordingBroker struct {
	mu        sync.Mutex
	releaseKeys []string
	reconcileKeys []string
	releaseResult string
	releaseErr   error
	reconcileResult string
	reconcileErr    error
}

func (b *recordingBroker) Release(_ context.Context, _ string, key string) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.releaseKeys = append(b.releaseKeys, key)
	return b.releaseResult, b.releaseErr
}
func (b *recordingBroker) Reconcile(_ context.Context, key string) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.reconcileKeys = append(b.reconcileKeys, key)
	return b.reconcileResult, b.reconcileErr
}

func newRetryService(t *testing.T, b customs.Broker) (*customs.RetryService, *storage.Store, domain.User, string) {
	t.Helper()
	f, _, d := setup(t)
	cs := customs.Service{Store: f.Store}
	if err := cs.Submit(context.Background(), f.User, d, "r"); err != nil {
		t.Fatal(err)
	}
	svc := &customs.RetryService{Store: f.Store, Broker: b}
	return svc, f.Store, f.User, d
}

func TestRetryReusesBrokerOperationKeyAfterTimeout(t *testing.T) {
	// First attempt: Release times out (transport error) and Reconcile cannot
	// confirm release, so the worker must surface the error WITHOUT minting a
	// new key. The second attempt must reuse the very same key.
	broker := &recordingBroker{
		releaseErr:      errors.New("context deadline exceeded"),
		reconcileErr:    errors.New("unknown"),
		reconcileResult: "",
	}
	svc, store, user, declID := newRetryService(t, broker)

	if err := svc.Run(context.Background(), user, declID); err == nil {
		t.Fatal("first timed-out attempt should surface an error")
	}

	// Second attempt: broker now succeeds.
	broker.releaseResult = "released"
	broker.releaseErr = nil
	if err := svc.Run(context.Background(), user, declID); err != nil {
		t.Fatalf("second attempt: %v", err)
	}

	// Exactly one key was minted and reused on both Release calls.
	if len(broker.releaseKeys) != 2 {
		t.Fatalf("expected 2 Release calls, got %d", len(broker.releaseKeys))
	}
	if broker.releaseKeys[0] == "" {
		t.Fatal("first release used an empty key")
	}
	if broker.releaseKeys[0] != broker.releaseKeys[1] {
		t.Fatalf("retry minted a new key: %q then %q", broker.releaseKeys[0], broker.releaseKeys[1])
	}

	// The persisted key matches the one sent to the broker, and the
	// declaration is released.
	var storedKey, status string
	if err := store.QueryRow(context.Background(),
		"SELECT broker_operation_key,status FROM customs_declarations WHERE id=?", declID).
		Scan(&storedKey, &status); err != nil {
		t.Fatal(err)
	}
	if storedKey != broker.releaseKeys[0] {
		t.Fatalf("persisted key %q != broker key %q", storedKey, broker.releaseKeys[0])
	}
	if status != "released" {
		t.Fatalf("status=%s want released", status)
	}
}

func TestRetryReconcilesViaSameKey(t *testing.T) {
	// First attempt times out, but Reconcile confirms release via the SAME key.
	broker := &recordingBroker{
		releaseErr:      errors.New("context deadline exceeded"),
		reconcileResult: "released",
	}
	svc, store, user, declID := newRetryService(t, broker)

	if err := svc.Run(context.Background(), user, declID); err != nil {
		t.Fatalf("reconciled release should succeed: %v", err)
	}
	if len(broker.releaseKeys) != 1 || len(broker.reconcileKeys) != 1 {
		t.Fatalf("release=%d reconcile=%d", len(broker.releaseKeys), len(broker.reconcileKeys))
	}
	if broker.releaseKeys[0] != broker.reconcileKeys[0] {
		t.Fatalf("reconcile used a different key than release: %q vs %q",
			broker.reconcileKeys[0], broker.releaseKeys[0])
	}
	var status string
	if err := store.QueryRow(context.Background(),
		"SELECT status FROM customs_declarations WHERE id=?", declID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "released" {
		t.Fatalf("status=%s want released", status)
	}
}
