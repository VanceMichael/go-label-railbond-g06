package query_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/query"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestManifestSearchAndStatuses(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c := f.Consignment(t, tr)
	var ref string
	if err := f.Store.QueryRow(context.Background(), "SELECT reference FROM consignments WHERE id=?", c).Scan(&ref); err != nil {
		t.Fatal(err)
	}
	s := query.ManifestQuery{Store: f.Store}
	found, err := s.Search(context.Background(), f.User, ref, 10)
	if err != nil || len(found) != 1 {
		t.Fatalf("%v %d", err, len(found))
	}
	statuses, err := s.Statuses(context.Background(), f.User)
	if err != nil || statuses["in_transit"] != 1 {
		t.Fatalf("%v %#v", err, statuses)
	}
}
