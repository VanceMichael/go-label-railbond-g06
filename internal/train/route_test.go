package train_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/train"
	"testing"
)

func TestRouteIncludesCheckpointNames(t *testing.T) {
	f := testkit.New(t)
	s := train.Service{Store: f.Store}
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO checkpoints(id,tenant_id,corridor_id,sequence_no,name) VALUES(?,?,?,?,?)", "cp", f.TenantID, f.CorridorID, 1, "Mohan"); err != nil {
		t.Fatal(err)
	}
	r, err := s.Route(context.Background(), f.User, f.CorridorID)
	if err != nil || len(r.Stops) != 1 {
		t.Fatalf("%v %#v", err, r)
	}
}
func TestRouteValidation(t *testing.T) {
	if err := train.ValidateRoute(train.Route{Origin: "A", Destination: "A"}); err == nil {
		t.Fatal("same endpoints accepted")
	}
}
