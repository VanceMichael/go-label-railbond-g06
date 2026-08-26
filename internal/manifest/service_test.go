package manifest_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/manifest"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestBuildAndValidateManifest(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	_ = f.Consignment(t, tr)
	s := manifest.Service{Store: f.Store}
	items, err := s.Build(context.Background(), f.User, tr)
	if err != nil || len(items) != 1 {
		t.Fatalf("%v %d", err, len(items))
	}
	if err := s.Validate(context.Background(), f.User, tr); err != nil {
		t.Fatal(err)
	}
}
func TestSortEntriesCopiesInput(t *testing.T) {
	in := []manifest.Entry{{Reference: "B", ConsignmentID: "2"}, {Reference: "A", ConsignmentID: "1"}}
	out := manifest.SortEntries(in)
	if out[0].Reference != "A" || in[0].Reference != "B" {
		t.Fatal(out)
	}
}
