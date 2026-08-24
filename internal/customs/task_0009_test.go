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

func TestTask0009BatchReleaseHasNoPartialCommit(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	cs := consignment.Service{Store: f.Store}
	a, _ := cs.Create(context.Background(), f.User, consignment.CreateInput{TrainID: tr, ContainerID: f.ContainerID, Reference: "a", Items: []consignment.Item{{SKU: "a", Description: "a", Quantity: 1}}}, "r")
	b, _ := cs.Create(context.Background(), f.User, consignment.CreateInput{TrainID: tr, ContainerID: f.ContainerID, Reference: "b", Items: []consignment.Item{{SKU: "b", Description: "b", Quantity: 1}}}, "r")
	s := customs.Service{Store: f.Store}
	da, _ := s.Create(context.Background(), f.User, a.ID)
	db, _ := s.Create(context.Background(), f.User, b.ID)
	_ = s.Submit(context.Background(), f.User, da.ID, "r")
	_ = s.Submit(context.Background(), f.User, db.ID, "r")
	_ = s.Hold(context.Background(), f.User, db.ID, "hold")
	_, err := (&customs.BatchService{Store: f.Store}).Release(context.Background(), f.User, []string{da.ID, db.ID}, "r")
	if !errors.Is(err, domain.ErrDeclarationHold) {
		t.Fatalf("got %v", err)
	}
	got, _ := s.Get(context.Background(), f.User, da.ID)
	if got.Status != "submitted" {
		t.Fatalf("partial release=%s", got.Status)
	}
}
