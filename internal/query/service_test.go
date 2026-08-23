package query_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/query"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
	"time"
)

func TestETAAndOverview(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	eta, err := (&query.ETAService{Store: f.Store}).Estimate(context.Background(), f.User, c)
	if err != nil || eta.Before(time.Now().UTC()) {
		t.Fatalf("%v %v", err, eta)
	}
	o, err := (&query.OverviewService{Store: f.Store}).Load(context.Background(), f.User)
	if err != nil || o.Consignments != 1 {
		t.Fatalf("%v %#v", err, o)
	}
}
func TestTimelineAndTenantFilter(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	_, err := (&query.TimelineRepository{Store: f.Store}).Load(context.Background(), f.User, c)
	if err == nil {
		t.Fatal("empty timeline should be missing")
	}
}
func TestRouteCacheCopiesSlices(t *testing.T) {
	c := query.NewRouteCache()
	c.Put("r", query.RouteSnapshot{Carrier: "A", Stops: []string{"K", "G"}})
	r, ok := c.Get("r")
	if !ok {
		t.Fatal("missing route")
	}
	r.Stops[0] = "mutated"
	again, _ := c.Get("r")
	if again.Stops[0] != "K" {
		t.Fatal("cache was polluted")
	}
}
func TestManifestPage(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	_ = f.Consignment(t, tr)
	p, err := (&query.ManifestQuery{Store: f.Store}).List(context.Background(), f.User, tr, "", "", 10)
	if err != nil || len(p.Items) != 1 || p.Total != 1 {
		t.Fatalf("%v %#v", err, p)
	}
}
