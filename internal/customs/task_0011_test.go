package customs_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/customs"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

type taskBroker11 struct{ release, reconcile int }

func (b *taskBroker11) Release(context.Context, string, string) (string, error) {
	b.release++
	return "", errors.New("response lost")
}
func (b *taskBroker11) Reconcile(context.Context, string) (string, error) {
	b.reconcile++
	return "released", nil
}
func TestTask0011CustomsRetryReconcilesAcceptedOperation(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c := f.Consignment(t, tr)
	d := f.Declaration(t, c, "submitted")
	b := &taskBroker11{}
	s := &customs.RetryService{Store: f.Store, Broker: b}
	if err := s.Run(context.Background(), f.User, d); err != nil {
		t.Fatal(err)
	}
	got, _ := (&customs.Service{Store: f.Store}).Get(context.Background(), f.User, d)
	if got.Status != "released" || b.release != 1 || b.reconcile != 1 {
		t.Fatalf("status=%s release=%d reconcile=%d", got.Status, b.release, b.reconcile)
	}
}
