package checkpoint_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/checkpoint"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestTask0019CheckpointProgressNeverRegresses(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c := f.Consignment(t, tr)
	cp2, cp1 := storage.NewID(), storage.NewID()
	for _, item := range []struct {
		id  string
		seq int
	}{{cp2, 2}, {cp1, 1}} {
		if _, err := f.Store.Exec(context.Background(), "INSERT INTO checkpoints(id,tenant_id,corridor_id,sequence_no,name) VALUES(?,?,?,?,?)", item.id, f.TenantID, f.CorridorID, item.seq, item.id); err != nil {
			t.Fatal(err)
		}
	}
	s := &checkpoint.Service{Store: f.Store}
	if _, err := s.RecordScan(context.Background(), f.User, c, cp2, "s2", "h2", "r"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RecordScan(context.Background(), f.User, c, cp1, "s1", "h1", "r"); err != nil {
		t.Fatal(err)
	}
	var progress int
	_ = f.Store.QueryRow(context.Background(), "SELECT current_checkpoint FROM consignments WHERE id=?", c).Scan(&progress)
	if progress != 2 {
		t.Fatalf("progress regressed to %d", progress)
	}
}
