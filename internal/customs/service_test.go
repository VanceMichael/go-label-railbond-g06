package customs_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/consignment"
	"github.com/VanceMichael/go-base-railbond-g06/internal/customs"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func setup(t *testing.T) (testkit.Fixture, string, string) {
	f := testkit.New(t)
	tr := f.Train(t)
	cs := consignment.Service{Store: f.Store}
	c, err := cs.Create(context.Background(), f.User, consignment.CreateInput{TrainID: tr, ContainerID: f.ContainerID, Reference: "customs", Items: []consignment.Item{{SKU: "x", Description: "x", Quantity: 1}}}, "r")
	if err != nil {
		t.Fatal(err)
	}
	d, err := customs.Service{Store: f.Store}.Create(context.Background(), f.User, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	return f, c.ID, d.ID
}
func TestSubmitRelease(t *testing.T) {
	f, _, d := setup(t)
	s := customs.Service{Store: f.Store}
	if err := s.Submit(context.Background(), f.User, d, "r"); err != nil {
		t.Fatal(err)
	}
	if err := s.Release(context.Background(), f.User, d, "r"); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(context.Background(), f.User, d)
	if err != nil || got.Status != "released" {
		t.Fatalf("%v %#v", err, got)
	}
}
func TestHoldPreservesState(t *testing.T) {
	f, _, d := setup(t)
	s := customs.Service{Store: f.Store}
	if err := s.Submit(context.Background(), f.User, d, "r"); err != nil {
		t.Fatal(err)
	}
	if err := s.Hold(context.Background(), f.User, d, "missing invoice"); err != nil {
		t.Fatal(err)
	}
	if err := s.Release(context.Background(), f.User, d, "r"); !errors.Is(err, domain.ErrDeclarationHold) {
		t.Fatal(err)
	}
}
func TestBatchIsAtomic(t *testing.T) {
	f, _, d := setup(t)
	tr := f.Train(t)
	cs := consignment.Service{Store: f.Store}
	c2, _ := cs.Create(context.Background(), f.User, consignment.CreateInput{TrainID: tr, ContainerID: f.ContainerID, Reference: "customs2", Items: []consignment.Item{{SKU: "y", Description: "y", Quantity: 1}}}, "r")
	d2, _ := customs.Service{Store: f.Store}.Create(context.Background(), f.User, c2.ID)
	s := customs.BatchService{Store: f.Store}
	if _, err := s.Release(context.Background(), f.User, []string{d, d2.ID}, "r"); err == nil {
		t.Fatal("draft batch released")
	}
	// Neither declaration should have moved: the batch must roll back fully.
	for _, id := range []string{d, d2.ID} {
		got, err := customs.Service{Store: f.Store}.Get(context.Background(), f.User, id)
		if err != nil || got.Status == "released" {
			t.Fatalf("batch leaked partial release for %s: %#v", id, got)
		}
	}
}

func TestBatchRollsBackOnHold(t *testing.T) {
	// Reproduces the reported scenario: a releaseable declaration precedes a
	// hold declaration in the batch. Without an atomic batch the first item
	// would be committed released while the second fails, leaving the batch
	// partially submitted.
	f, _, d := setup(t)
	tr := f.Train(t)
	cs := consignment.Service{Store: f.Store}
	c2, _ := cs.Create(context.Background(), f.User, consignment.CreateInput{TrainID: tr, ContainerID: f.ContainerID, Reference: "customs2", Items: []consignment.Item{{SKU: "y", Description: "y", Quantity: 1}}}, "r")
	d2, _ := customs.Service{Store: f.Store}.Create(context.Background(), f.User, c2.ID)

	single := customs.Service{Store: f.Store}
	if err := single.Submit(context.Background(), f.User, d, "r"); err != nil {
		t.Fatal(err)
	}
	if err := single.Submit(context.Background(), f.User, d2.ID, "r"); err != nil {
		t.Fatal(err)
	}
	if err := single.Hold(context.Background(), f.User, d2.ID, "missing invoice"); err != nil {
		t.Fatal(err)
	}

	s := customs.BatchService{Store: f.Store}
	n, err := s.Release(context.Background(), f.User, []string{d, d2.ID}, "r")
	if !errors.Is(err, domain.ErrDeclarationHold) {
		t.Fatalf("want hold error, got %v", err)
	}
	if n != 0 {
		t.Fatalf("want 0 released on rollback, got %d", n)
	}
	got, err := single.Get(context.Background(), f.User, d)
	if err != nil || got.Status != "submitted" {
		t.Fatalf("first item leaked to released: %#v", got)
	}
}
