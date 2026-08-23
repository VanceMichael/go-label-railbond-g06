package testkit

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/auth"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"path/filepath"
	"testing"
	"time"
)

type Fixture struct {
	Store                                              *storage.Store
	User                                               domain.User
	Auth                                               auth.Service
	TenantID, CorridorID, TrainID, ContainerID, SlotID string
}

func New(t *testing.T) Fixture {
	t.Helper()
	st, err := storage.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	if err := st.Migrate(context.Background(), filepath.Join("..", "..", "migrations", "001_initial.sql")); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tenantID := storage.NewID()
	if _, err := st.Exec(context.Background(), "INSERT INTO tenants(id,name,timezone,created_at) VALUES(?,?,?,?)", tenantID, "Test Tenant", "Asia/Shanghai", now); err != nil {
		t.Fatal(err)
	}
	uid := storage.NewID()
	if _, err := st.Exec(context.Background(), "INSERT INTO users(id,tenant_id,email,role,password_hash,created_at) VALUES(?,?,?,?,?,?)", uid, tenantID, "operator@example.test", "operator", "test", now); err != nil {
		t.Fatal(err)
	}
	corr := storage.NewID()
	if _, err := st.Exec(context.Background(), "INSERT INTO corridors(id,tenant_id,name,origin,destination,timezone,created_at) VALUES(?,?,?,?,?,?,?)", corr, tenantID, "Kunming-Guiyang", "Kunming", "Guiyang", "Asia/Shanghai", now); err != nil {
		t.Fatal(err)
	}
	cont := storage.NewID()
	if _, err := st.Exec(context.Background(), "INSERT INTO containers(id,tenant_id,code,status,version,created_at) VALUES(?,?,?,?,?,?)", cont, tenantID, "RBXU000001", "available", 1, now); err != nil {
		t.Fatal(err)
	}
	slot := storage.NewID()
	if _, err := st.Exec(context.Background(), "INSERT INTO warehouse_slots(id,tenant_id,code,zone,status,version) VALUES(?,?,?,?,?,1)", slot, tenantID, "A-01", "bonded", "available"); err != nil {
		t.Fatal(err)
	}
	user := domain.User{ID: uid, TenantID: tenantID, Email: "operator@example.test", Role: "operator"}
	return Fixture{Store: st, User: user, Auth: auth.Service{Store: st}, TenantID: tenantID, CorridorID: corr, ContainerID: cont, SlotID: slot}
}
func (f Fixture) Admin() domain.User   { u := f.User; u.Role = "admin"; return u }
func (f Fixture) Finance() domain.User { u := f.User; u.Role = "finance"; return u }
func (f Fixture) Train(t *testing.T) string {
	t.Helper()
	now := time.Now().UTC().Add(24 * time.Hour)
	id := storage.NewID()
	slot := storage.NewID()
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO rail_slots(id,tenant_id,corridor_id,starts_at,ends_at,status,version) VALUES(?,?,?,?,?,?,1)", slot, f.TenantID, f.CorridorID, now.Add(-time.Hour).Format(time.RFC3339Nano), now.Add(time.Hour).Format(time.RFC3339Nano), "held"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO trains(id,tenant_id,corridor_id,number,status,capacity,slot_id,departure_at,created_at) VALUES(?,?,?,?,?,?,?,?,?)", id, f.TenantID, f.CorridorID, "RB"+id, "published", 10, slot, now.Format(time.RFC3339Nano), ts); err != nil {
		t.Fatal(err)
	}
	return id
}
func (f Fixture) Consignment(t *testing.T, train string) string {
	t.Helper()
	id := storage.NewID()
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO consignments(id,tenant_id,train_id,container_id,reference,status,created_at) VALUES(?,?,?,?,?,?,?)", id, f.TenantID, train, f.ContainerID, "REF-"+id[:8], "in_transit", time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO consignment_items(id,consignment_id,sku,description,quantity,declared_value) VALUES(?,?,?,?,?,?)", storage.NewID(), id, "COFFEE", "Coffee beans", 2, 300); err != nil {
		t.Fatal(err)
	}
	return id
}
func (f Fixture) Declaration(t *testing.T, consignment string, status string) string {
	t.Helper()
	id := storage.NewID()
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO customs_declarations(id,tenant_id,consignment_id,status,version) VALUES(?,?,?,?,1)", id, f.TenantID, consignment, status); err != nil {
		t.Fatal(err)
	}
	return id
}
