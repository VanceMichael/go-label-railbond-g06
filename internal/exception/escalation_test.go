package exception_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/exception"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
	"time"
)

func TestEscalateAndReopen(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	s := exception.Service{Store: f.Store}
	id, _ := s.Open(context.Background(), f.User, c, "delay")
	if err := s.Escalate(context.Background(), f.User, id, "ops", time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := s.Resolve(context.Background(), f.User, id, "alternate", "r"); err != nil {
		t.Fatal(err)
	}
	if err := s.Reopen(context.Background(), f.User, id); err != nil {
		t.Fatal(err)
	}
}
