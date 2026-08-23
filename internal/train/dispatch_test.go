package train_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/train"
	"testing"
	"time"
)

func TestDispatchReadyOnlyMovesReleasedCargo(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c := f.Consignment(t, tr)
	if _, err := f.Store.Exec(context.Background(), "UPDATE consignments SET status='booked' WHERE id=?", c); err != nil {
		t.Fatal(err)
	}
	d := f.Declaration(t, c, "released")
	_ = d
	r, err := (&train.Service{Store: f.Store}).DispatchReady(context.Background(), f.User, tr)
	if err != nil || r.Moved != 1 {
		t.Fatalf("%v %#v", err, r)
	}
}
func TestPastDepartureRejected(t *testing.T) {
	f := testkit.New(t)
	if err := (&train.Service{Store: f.Store}).DepartAt(context.Background(), f.User, f.Train(t), fTime()); err == nil {
		t.Fatal("past departure accepted")
	}
}
func fTime() time.Time { return time.Now().UTC().Add(-time.Hour) }
