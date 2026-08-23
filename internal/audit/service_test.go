package audit_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/audit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestAuditRecentAndObject(t *testing.T) {
	f := testkit.New(t)
	s := audit.Service{Store: f.Store}
	if err := s.Write(context.Background(), f.User, "train.created", "train", "t1", "success", "req", "created"); err != nil {
		t.Fatal(err)
	}
	recent, err := s.Recent(context.Background(), f.User, 10)
	if err != nil || len(recent) != 1 {
		t.Fatalf("%v %d", err, len(recent))
	}
	obj, err := s.ForObject(context.Background(), f.User, "train", "t1")
	if err != nil || len(obj) != 1 {
		t.Fatalf("%v %d", err, len(obj))
	}
}
func TestAuditHookFails(t *testing.T) {
	f := testkit.New(t)
	f.Store.Hooks.BeforeAudit = func(context.Context) error { return context.Canceled }
	if err := (&audit.Service{Store: f.Store}).Write(context.Background(), f.User, "x", "x", "x", "failed", "r", "d"); err == nil {
		t.Fatal("audit hook ignored")
	}
}
