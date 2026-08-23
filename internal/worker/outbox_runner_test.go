package worker_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/worker"
	"testing"
	"time"
)

type publisher struct{ n int }

func (p *publisher) Publish(context.Context, string, string, string) error { p.n++; return nil }
func TestOutboxRunnerPublishes(t *testing.T) {
	f := testkit.New(t)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO outbox_messages(id,tenant_id,topic,aggregate_id,payload,status,available_at,created_at) VALUES(?,?,?,?,?,?,?,?)", "run1", f.TenantID, "topic", "aggregate", "body", "pending", now, now); err != nil {
		t.Fatal(err)
	}
	p := &publisher{}
	r := worker.OutboxRunner{Store: f.Store, Publisher: p, Owner: "runner", Epoch: 1}
	if err := r.RunOnce(context.Background(), f.TenantID, "run1"); err != nil || p.n != 1 {
		t.Fatalf("%v %d", err, p.n)
	}
}
