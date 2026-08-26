package query_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/query"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestTask0016TimelineCannotCrossTenant(t *testing.T) {
	f := testkit.New(t)
	other := storage.NewID()
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO tenants(id,name,timezone,created_at) VALUES(?,?,?,datetime('now'))", other, "Other", "UTC"); err != nil {
		t.Fatal(err)
	}
	corr := storage.NewID()
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO corridors(id,tenant_id,name,origin,destination,timezone,created_at) VALUES(?,?,?,?,?,?,datetime('now'))", corr, other, "Other", "A", "B", "UTC"); err != nil {
		t.Fatal(err)
	}
	tr := storage.NewID()
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO trains(id,tenant_id,corridor_id,number,status,capacity,departure_at,created_at) VALUES(?,?,?,?,?,?,datetime('now'),datetime('now'))", tr, other, corr, "OTHER-16", "published", 2); err != nil {
		t.Fatal(err)
	}
	cont := storage.NewID()
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO containers(id,tenant_id,code,status,created_at) VALUES(?,?,?,?,datetime('now'))", cont, other, "OTHER-C", "available"); err != nil {
		t.Fatal(err)
	}
	c := storage.NewID()
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO consignments(id,tenant_id,train_id,container_id,reference,status,created_at) VALUES(?,?,?,?,?,?,datetime('now'))", c, other, tr, cont, "OTHER-CARGO", "in_transit"); err != nil {
		t.Fatal(err)
	}
	cp := storage.NewID()
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO checkpoints(id,tenant_id,corridor_id,sequence_no,name) VALUES(?,?,?,?,?)", cp, other, corr, 1, "Other Border"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO checkpoint_events(id,tenant_id,checkpoint_id,consignment_id,scanner_id,evidence_hash,observed_at,created_at) VALUES(?,?,?,?,?,?,datetime('now'),datetime('now'))", storage.NewID(), other, cp, c, "scanner", "secret"); err != nil {
		t.Fatal(err)
	}
	items, err := (&query.TimelineRepository{Store: f.Store}).Load(context.Background(), f.User, c)
	if err == nil || len(items) != 0 {
		t.Fatalf("cross tenant timeline leaked items=%d err=%v", len(items), err)
	}
}
