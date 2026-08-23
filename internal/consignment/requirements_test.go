package consignment_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/consignment"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestRequirementsReflectEvidence(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c := f.Consignment(t, tr)
	s := consignment.Service{Store: f.Store}
	items, err := s.Requirements(context.Background(), f.User, c)
	if err != nil || len(items) != 3 {
		t.Fatalf("%v %d", err, len(items))
	}
	if err := s.CanDeliver(context.Background(), f.User, c); err == nil {
		t.Fatal("incomplete cargo deliverable")
	}
}
