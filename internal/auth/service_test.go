package auth_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/auth"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
	"time"
)

func TestLoginLogoutAndExpiry(t *testing.T) {
	f := testkit.New(t)
	s := auth.Service{Store: f.Store, Now: func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }}
	if _, err := s.CreateUser(context.Background(), f.TenantID, "alice@example.test", "customs", "secret"); err != nil {
		t.Fatal(err)
	}
	u, token, err := s.Login(context.Background(), f.TenantID, "alice@example.test", "secret")
	if err != nil || u.Role != "customs" {
		t.Fatalf("login: %v %#v", err, u)
	}
	if _, err := s.Authenticate(context.Background(), token); err != nil {
		t.Fatal(err)
	}
	if err := s.Logout(context.Background(), token); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Authenticate(context.Background(), token); err == nil {
		t.Fatal("revoked token accepted")
	}
}
func TestRoleGuard(t *testing.T) {
	if err := domain.RequireRole(domain.User{Role: "customs"}, "finance"); err == nil {
		t.Fatal("role accepted")
	}
}
func TestWrongPasswordRejected(t *testing.T) {
	f := testkit.New(t)
	s := auth.Service{Store: f.Store}
	if _, err := s.CreateUser(context.Background(), f.TenantID, "x@example.test", "operator", "right"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Login(context.Background(), f.TenantID, "x@example.test", "wrong"); err == nil {
		t.Fatal("wrong password accepted")
	}
}
