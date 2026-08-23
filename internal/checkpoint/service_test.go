package checkpoint_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/checkpoint"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func setup(t *testing.T) (testkit.Fixture, string, string) {
	f := testkit.New(t)
	tr := f.Train(t)
	c := f.Consignment(t, tr)
	cp := storage.NewID()
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO checkpoints(id,tenant_id,corridor_id,sequence_no,name,required_inspection) VALUES(?,?,?,?,?,0)", cp, f.TenantID, f.CorridorID, 1, "border", 0); err != nil {
		t.Fatal(err)
	}
	return f, c, cp
}
func TestScanAdvancesAndLists(t *testing.T) {
	f, c, cp := setup(t)
	s := checkpoint.Service{Store: f.Store}
	got, err := s.RecordScan(context.Background(), f.User, c, cp, "scanner-1", "hash-1", "r")
	if err != nil || got.EvidenceHash != "hash-1" {
		t.Fatalf("%v %#v", err, got)
	}
	events, err := s.Events(context.Background(), f.User, c)
	if err != nil || len(events) != 1 {
		t.Fatalf("%v %d", err, len(events))
	}
}
func TestDuplicateScanRejected(t *testing.T) {
	f, c, cp := setup(t)
	s := checkpoint.Service{Store: f.Store}
	_, _ = s.RecordScan(context.Background(), f.User, c, cp, "s", "h", "r")
	_, err := s.RecordScan(context.Background(), f.User, c, cp, "s", "h2", "r")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatal(err)
	}
}
func TestRequiredInspection(t *testing.T) {
	f, c, cp := setup(t)
	if _, err := f.Store.Exec(context.Background(), "UPDATE checkpoints SET required_inspection=1 WHERE id=?", cp); err != nil {
		t.Fatal(err)
	}
	s := checkpoint.Service{Store: f.Store}
	if err := s.CreateInspection(context.Background(), f.User, c, cp); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RecordScan(context.Background(), f.User, c, cp, "s", "h", "r"); err == nil {
		t.Fatal("pending inspection accepted")
	}
	if err := s.PassInspection(context.Background(), f.User, c, cp); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RecordScan(context.Background(), f.User, c, cp, "s", "h", "r"); err != nil {
		t.Fatal(err)
	}
}
