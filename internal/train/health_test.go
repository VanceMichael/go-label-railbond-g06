package train_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/train"
	"testing"
	"time"
)

func TestTrainHealth(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	h, err := (&train.Service{Store: f.Store}).Health(context.Background(), f.User, tr)
	if err != nil || h.Capacity != 10 {
		t.Fatalf("%v %#v", err, h)
	}
	if ok, err := (&train.Service{Store: f.Store}).ReadyForDeparture(context.Background(), f.User, tr); err != nil || ok {
		t.Fatalf("%v %v", err, ok)
	}
}
func TestScheduleSummary(t *testing.T) {
	f := testkit.New(t)
	s := &train.Service{Store: f.Store}
	if rows, err := s.ScheduleSummary(context.Background(), f.User, time.Now().UTC().Add(-time.Hour), time.Now().UTC().Add(48*time.Hour)); err != nil || rows == nil {
		t.Fatalf("%v", err)
	}
}
