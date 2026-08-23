package exception_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/exception"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestOpenResolveException(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	s := exception.Service{Store: f.Store}
	id, err := s.Open(context.Background(), f.User, c, "border delay")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Resolve(context.Background(), f.User, id, "alternate", "r"); err != nil {
		t.Fatal(err)
	}
	status, _, err := s.Get(context.Background(), f.User, id)
	if err != nil || status != "resolved" {
		t.Fatalf("%v %s", err, status)
	}
}
func TestExceptionReadIsTenantScoped(t *testing.T) {
	f := testkit.New(t)
	s := exception.Service{Store: f.Store}
	if _, _, err := s.Get(context.Background(), f.User, "missing"); err == nil {
		t.Fatal("missing exception returned")
	}
}
