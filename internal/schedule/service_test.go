package schedule_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/schedule"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
	"time"
)

func TestWindowConflict(t *testing.T) {
	f := testkit.New(t)
	s := schedule.Service{Store: f.Store}
	start := time.Now().UTC().Add(time.Hour)
	one, err := s.ReserveWindow(context.Background(), f.User, f.CorridorID, start, start.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ReserveWindow(context.Background(), f.User, f.CorridorID, start.Add(30*time.Minute), start.Add(90*time.Minute)); err == nil {
		t.Fatal("overlapping window accepted")
	}
	if err := s.ReleaseWindow(context.Background(), f.User, one); err != nil {
		t.Fatal(err)
	}
}
func TestOpenWindow(t *testing.T) {
	s := schedule.Service{}
	now := time.Now()
	if !s.IsOpen(now.Add(-time.Minute), now.Add(time.Minute), now) {
		t.Fatal("window should be open")
	}
}
