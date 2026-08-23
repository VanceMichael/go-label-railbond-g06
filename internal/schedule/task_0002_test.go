package schedule_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/schedule"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
	"time"
)

func TestTask0002RepeatedReleaseRejectsStaleWindow(t *testing.T) {
	f := testkit.New(t)
	s := schedule.Service{Store: f.Store}
	start := time.Now().UTC().Add(time.Hour)
	id, err := s.ReserveWindow(context.Background(), f.User, f.CorridorID, start, start.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseWindow(context.Background(), f.User, id); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseWindow(context.Background(), f.User, id); err == nil {
		t.Fatal("stale release was accepted")
	}
}
