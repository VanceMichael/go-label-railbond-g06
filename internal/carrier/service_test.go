package carrier_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/carrier"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
	"time"
)

type fakeClient struct{}

func (fakeClient) Accept(context.Context, string, string) (carrier.Receipt, error) {
	return carrier.Receipt{Status: "accepted", ProviderKey: "p", ReceivedAt: time.Now()}, nil
}
func (fakeClient) Cancel(context.Context, string, string) error { return nil }
func TestCarrierAcceptAndCancel(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c := f.Consignment(t, tr)
	if _, err := f.Store.Exec(context.Background(), "UPDATE consignments SET status='booked' WHERE id=?", c); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO route_assignments(id,tenant_id,consignment_id,carrier,status,next_attempt_at) VALUES(?,?,?,?,?,datetime('now'))", "ca1", f.TenantID, c, "carrier", "assigned"); err != nil {
		t.Fatal(err)
	}
	s := carrier.Service{Store: f.Store, Client: fakeClient{}}
	if _, err := s.Accept(context.Background(), f.User, "ca1", "provider"); err != nil {
		t.Fatal(err)
	}
	if err := s.Cancel(context.Background(), f.User, "ca1", "provider"); err != nil {
		t.Fatal(err)
	}
}
