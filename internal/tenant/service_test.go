package tenant_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/tenant"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestCorridorAndWarehouseSetup(t *testing.T) {
	f := testkit.New(t)
	s := tenant.Service{Store: f.Store}
	id, err := s.CreateCorridor(context.Background(), f.Admin(), "China-Laos", "Kunming", "Vientiane", "Asia/Shanghai")
	if err != nil || id == "" {
		t.Fatal(err)
	}
	slot, err := s.AddWarehouseSlot(context.Background(), f.Admin(), "B-01", "bonded")
	if err != nil || slot == "" {
		t.Fatal(err)
	}
	cp, err := s.AddCheckpoint(context.Background(), f.Admin(), f.CorridorID, 2, "Mohan", true)
	if err != nil || cp == "" {
		t.Fatal(err)
	}
}
func TestTenantRoleRestriction(t *testing.T) {
	f := testkit.New(t)
	if _, err := (&tenant.Service{Store: f.Store}).CreateCorridor(context.Background(), f.User, "x", "a", "b", "UTC"); err != nil {
		t.Fatal(err)
	}
}
