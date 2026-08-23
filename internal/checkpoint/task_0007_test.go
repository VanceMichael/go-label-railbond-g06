package checkpoint_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/checkpoint"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestTask0007CancelledScanDoesNotPersistEvidence(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c := f.Consignment(t, tr)
	cp := storage.NewID()
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO checkpoints(id,tenant_id,corridor_id,sequence_no,name) VALUES(?,?,?,?,?)", cp, f.TenantID, f.CorridorID, 1, "Mohan"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (&checkpoint.Service{Store: f.Store}).RecordScan(ctx, f.User, c, cp, "scanner", "hash", "r")
	if err == nil {
		t.Fatal("cancelled scan succeeded")
	}
	var n int
	if err := f.Store.QueryRow(context.Background(), "SELECT COUNT(*) FROM checkpoint_events WHERE consignment_id=?", c).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("cancelled scan persisted %d events", n)
	}
}
