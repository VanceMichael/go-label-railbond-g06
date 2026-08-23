package idempotency_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/idempotency"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestExecuteReplaysResponse(t *testing.T) {
	f := testkit.New(t)
	s := idempotency.Service{Store: f.Store}
	calls := 0
	op := func(_ *storage.Tx) (int, string, error) { calls++; return 201, "body", nil }
	r, err := s.Execute(context.Background(), f.User, "key", "POST", "/resource", op)
	if err != nil || r.Replay || r.Status != 201 {
		t.Fatalf("%v %#v", err, r)
	}
	r, err = s.Execute(context.Background(), f.User, "key", "POST", "/resource", op)
	if err != nil || !r.Replay || r.Body != "body" || calls != 1 {
		t.Fatalf("%v %#v calls=%d", err, r, calls)
	}
}
func TestKeyMismatchRejected(t *testing.T) {
	f := testkit.New(t)
	s := idempotency.Service{Store: f.Store}
	op := func(_ *storage.Tx) (int, string, error) { return 200, "ok", nil }
	_, _ = s.Execute(context.Background(), f.User, "same", "POST", "/a", op)
	if _, err := s.Execute(context.Background(), f.User, "same", "POST", "/b", op); err == nil {
		t.Fatal("mismatch accepted")
	}
	if _, err := s.Count(context.Background(), domain.User{TenantID: f.TenantID}); err != nil {
		t.Fatal(err)
	}
}
