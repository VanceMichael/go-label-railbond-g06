package settlement_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/consignment"
	"github.com/VanceMichael/go-base-railbond-g06/internal/settlement"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
	"time"
)

func TestQuoteAndPersist(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c, _ := (&consignment.Service{Store: f.Store}).Create(context.Background(), f.User, consignment.CreateInput{TrainID: tr, ContainerID: f.ContainerID, Reference: "quote", Items: []consignment.Item{{SKU: "a", Description: "a", Quantity: 2, DeclaredValue: 10}}}, "r")
	if _, err := f.Store.Exec(context.Background(), "UPDATE consignments SET status='delivered' WHERE id=?", c.ID); err != nil {
		t.Fatal(err)
	}
	s := &settlement.Service{Store: f.Store}
	q, err := s.Quote(context.Background(), f.User, c.ID, time.Now().UTC().Add(time.Hour))
	if err != nil || q.Total == 0 {
		t.Fatalf("%v %#v", err, q)
	}
	if err := s.PersistQuote(context.Background(), f.User, c.ID, q, "r"); err != nil {
		t.Fatal(err)
	}
}
