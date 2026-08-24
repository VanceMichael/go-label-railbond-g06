package customs_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/consignment"
	"github.com/VanceMichael/go-base-railbond-g06/internal/customs"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestTask0004SubmitAuditFailureLeavesDraft(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c, err := (&consignment.Service{Store: f.Store}).Create(context.Background(), f.User, consignment.CreateInput{TrainID: tr, ContainerID: f.ContainerID, Reference: "customs-audit", Items: []consignment.Item{{SKU: "x", Description: "x", Quantity: 1}}}, "r")
	if err != nil {
		t.Fatal(err)
	}
	d, err := (&customs.Service{Store: f.Store}).Create(context.Background(), f.User, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	f.Store.Hooks.BeforeAudit = func(context.Context) error { return errors.New("audit down") }
	if err := (&customs.Service{Store: f.Store}).Submit(context.Background(), f.User, d.ID, "r"); err == nil {
		t.Fatal("submit succeeded")
	}
	got, _ := (&customs.Service{Store: f.Store}).Get(context.Background(), f.User, d.ID)
	if got.Status != "draft" {
		t.Fatalf("status=%s", got.Status)
	}
}
