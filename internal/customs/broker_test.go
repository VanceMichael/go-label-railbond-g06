package customs_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/customs"
	"testing"
)

func TestBrokerReconcile(t *testing.T) {
	s, err := customs.ReconcileBroker(context.Background(), customs.StaticGateway{Value: customs.BrokerStatus{OperationKey: "k", Status: "released"}}, "k")
	if err != nil || s.Status != "released" {
		t.Fatalf("%v %#v", err, s)
	}
}
