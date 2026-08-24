package exception_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/exception"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestOpenResolveException(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	s := exception.Service{Store: f.Store}
	id, err := s.Open(context.Background(), f.User, c, "border delay")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Resolve(context.Background(), f.User, id, "alternate", "r"); err != nil {
		t.Fatal(err)
	}
	status, _, err := s.Get(context.Background(), f.User, id)
	if err != nil || status != "resolved" {
		t.Fatalf("%v %s", err, status)
	}
}
func TestExceptionReadIsTenantScoped(t *testing.T) {
	f := testkit.New(t)
	s := exception.Service{Store: f.Store}
	if _, _, err := s.Get(context.Background(), f.User, "missing"); err == nil {
		t.Fatal("missing exception returned")
	}
}
func TestResolveRollsBackOnOutboxFailure(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	s := exception.Service{Store: f.Store}
	id, err := s.Open(context.Background(), f.User, c, "border delay")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Store.Exec(context.Background(), "UPDATE containers SET lease_owner=?,lease_token=?,lease_until=? WHERE id=?", "owner", "tok", "9999-01-01T00:00:00Z", f.ContainerID); err != nil {
		t.Fatal(err)
	}
	f.Store.Hooks.BeforeOutbox = func(context.Context) error { return domain.ErrConflict }
	err = s.Resolve(context.Background(), f.User, id, "alternate", "r")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict error, got %v", err)
	}
	status, _, err := s.Get(context.Background(), f.User, id)
	if err != nil || status != "open" {
		t.Fatalf("exception should remain open on outbox failure: %v %s", err, status)
	}
	var leaseOwner string
	if err := f.Store.QueryRow(context.Background(), "SELECT lease_owner FROM containers WHERE id=?", f.ContainerID).Scan(&leaseOwner); err != nil {
		t.Fatal(err)
	}
	if leaseOwner != "owner" {
		t.Fatalf("container lease should be retained on outbox failure: %q", leaseOwner)
	}
	n, _ := f.Store.Count(context.Background(), "outbox_messages", f.TenantID)
	if n != 0 {
		t.Fatalf("expected no outbox messages on rollback, got %d", n)
	}
}
