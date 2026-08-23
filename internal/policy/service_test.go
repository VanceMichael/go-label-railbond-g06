package policy_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/policy"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestPolicyChangesWithLifecycle(t *testing.T) {
	f := testkit.New(t)
	id := f.Consignment(t, f.Train(t))
	p, err := (&policy.Evaluator{Store: f.Store}).ForConsignment(context.Background(), f.User, id)
	if err != nil || p.RequiredDocuments != 2 {
		t.Fatalf("%v %#v", err, p)
	}
}
func TestDeparturePolicy(t *testing.T) {
	f := testkit.New(t)
	ok, err := (&policy.Evaluator{Store: f.Store}).MayDepart(context.Background(), f.User, f.Train(t))
	if err != nil || !ok {
		t.Fatalf("%v %v", err, ok)
	}
}
