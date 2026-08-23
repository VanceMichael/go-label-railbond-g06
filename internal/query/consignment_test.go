package query_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/query"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestConsignmentViewJoinsTenantObjects(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c := f.Consignment(t, tr)
	v, err := (&query.OverviewService{Store: f.Store}).Consignment(context.Background(), f.User, c)
	if err != nil || v.ID != c || v.TrainNumber == "" {
		t.Fatalf("%v %#v", err, v)
	}
}
