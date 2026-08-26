package outbox_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/outbox"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
)

func TestTask0012StaleFailureCannotRequeueReclaimedEvent(t *testing.T) {
	f := testkit.New(t)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO outbox_messages(id,tenant_id,topic,aggregate_id,payload,status,lease_owner,lease_epoch,lease_until,available_at,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)", "event-12", f.TenantID, "topic", "agg", "body", "sending", "worker-b", 2, time.Now().UTC().Add(time.Minute).Format(time.RFC3339Nano), now, now); err != nil {
		t.Fatal(err)
	}

	s := outbox.Service{Store: f.Store}
	if err := s.Fail(context.Background(), f.User, "event-12", "worker-a", 1, "stale network error"); !errors.Is(err, domain.ErrLeaseLost) {
		t.Fatalf("stale worker error = %v, want lease lost", err)
	}

	var status, owner string
	if err := f.Store.QueryRow(context.Background(), "SELECT status,lease_owner FROM outbox_messages WHERE id='event-12'").Scan(&status, &owner); err != nil {
		t.Fatal(err)
	}
	if status != "sending" || owner != "worker-b" {
		t.Fatalf("stale failure changed current lease: status=%s owner=%s", status, owner)
	}

	if err := s.Fail(context.Background(), f.User, "event-12", "worker-b", 2, "publisher unavailable"); err != nil {
		t.Fatalf("current owner could not requeue event: %v", err)
	}
	var releasedOwner sql.NullString
	var lastError string
	if err := f.Store.QueryRow(context.Background(), "SELECT status,lease_owner,last_error FROM outbox_messages WHERE id='event-12'").Scan(&status, &releasedOwner, &lastError); err != nil {
		t.Fatal(err)
	}
	if status != "pending" || releasedOwner.Valid || lastError != "publisher unavailable" {
		t.Fatalf("valid failure did not release for retry: status=%s owner=%v error=%q", status, releasedOwner, lastError)
	}
}
