package query_test

import (
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/query"
	"testing"
)

func TestTask0017ForUserSnapshotCannotPolluteCache(t *testing.T) {
	c := query.NewRouteCache()
	c.Put("route-17", query.RouteSnapshot{Carrier: "carrier", Stops: []string{"Kunming", "Guiyang"}})
	r, ok := c.ForUser(domain.User{TenantID: "tenant"}, "route-17")
	if !ok {
		t.Fatal("missing route")
	}
	r.Stops[0] = "mutated"
	again, _ := c.Get("route-17")
	if again.Stops[0] != "Kunming" {
		t.Fatalf("cache polluted with %v", again.Stops)
	}
}
