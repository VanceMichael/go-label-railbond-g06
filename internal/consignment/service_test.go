package consignment_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/consignment"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestCreatePersistsItemsAndAudit(t *testing.T) {
	f := testkit.New(t)
	train := f.Train(t)
	s := consignment.Service{Store: f.Store}
	r, err := s.Create(context.Background(), f.User, consignment.CreateInput{TrainID: train, ContainerID: f.ContainerID, Reference: "RB-REF-1", Items: []consignment.Item{{SKU: "S1", Description: "tea", Quantity: 2, DeclaredValue: 10}}}, "req-1")
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != "booked" {
		t.Fatal(r)
	}
	items, err := s.Items(context.Background(), f.User, r.ID)
	if err != nil || len(items) != 1 {
		t.Fatalf("%v %d", err, len(items))
	}
	var audits int
	if err := f.Store.QueryRow(context.Background(), "SELECT COUNT(*) FROM audit_events WHERE object_id=?", r.ID).Scan(&audits); err != nil || audits != 1 {
		t.Fatalf("%v %d", err, audits)
	}
}
func TestCreateRejectsEmptyItems(t *testing.T) {
	f := testkit.New(t)
	s := consignment.Service{Store: f.Store}
	_, err := s.Create(context.Background(), f.User, consignment.CreateInput{TrainID: f.Train(t), ContainerID: f.ContainerID, Reference: "empty"}, "r")
	if !errors.Is(err, domain.ErrInvalidState) {
		t.Fatal(err)
	}
}
func TestAdvanceStateMachine(t *testing.T) {
	f := testkit.New(t)
	train := f.Train(t)
	s := consignment.Service{Store: f.Store}
	r, err := s.Create(context.Background(), f.User, consignment.CreateInput{TrainID: train, ContainerID: f.ContainerID, Reference: "advance", Items: []consignment.Item{{SKU: "x", Description: "x", Quantity: 1}}}, "r")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.MarkInTransit(context.Background(), f.User, r.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.Advance(context.Background(), f.User, r.ID, domain.ConsignmentArchived); !errors.Is(err, domain.ErrInvalidState) {
		t.Fatal(err)
	}
}
