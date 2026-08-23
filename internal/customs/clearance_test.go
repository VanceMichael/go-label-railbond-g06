package customs_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/consignment"
	"github.com/VanceMichael/go-base-railbond-g06/internal/customs"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestClearAndReopenDeclaration(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c, _ := (&consignment.Service{Store: f.Store}).Create(context.Background(), f.User, consignment.CreateInput{TrainID: tr, ContainerID: f.ContainerID, Reference: "clear", Items: []consignment.Item{{SKU: "a", Description: "a", Quantity: 1}}}, "r")
	d, _ := (&customs.Service{Store: f.Store}).Create(context.Background(), f.User, c.ID)
	if err := (&customs.Service{Store: f.Store}).Submit(context.Background(), f.User, d.ID, "r"); err != nil {
		t.Fatal(err)
	}
	if _, err := (&customs.Service{Store: f.Store}).Clear(context.Background(), f.User, d.ID, "r"); err != nil {
		t.Fatal(err)
	}
	if err := (&customs.Service{Store: f.Store}).Reopen(context.Background(), f.User, d.ID, "new hold"); err != nil {
		t.Fatal(err)
	}
}
