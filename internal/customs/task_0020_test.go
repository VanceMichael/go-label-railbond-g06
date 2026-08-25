package customs_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/customs"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

type taskBroker20 struct{ keys []string }

func (b *taskBroker20) Release(_ context.Context, _ string, key string) (string, error) {
	b.keys = append(b.keys, key)
	return "", errors.New("timeout")
}
func (b *taskBroker20) Reconcile(context.Context, string) (string, error) {
	return "unknown", errors.New("unknown")
}

func TestTask0020RetryReusesBrokerOperationKey(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	d := f.Declaration(t, c, "submitted")
	b := &taskBroker20{}
	s := &customs.RetryService{Store: f.Store, Broker: b}
	_ = s.Run(context.Background(), f.User, d)
	_ = s.Run(context.Background(), f.User, d)
	if len(b.keys) != 2 || b.keys[0] != b.keys[1] {
		t.Fatalf("operation keys=%v", b.keys)
	}
}
