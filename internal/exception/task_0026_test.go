package exception_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/VanceMichael/go-base-railbond-g06/internal/exception"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
)

func TestTask0026ResolveAuditFailurePreservesExceptionAndLease(t *testing.T) {
	f := testkit.New(t)
	consignmentID := f.Consignment(t, f.Train(t))
	service := exception.Service{Store: f.Store}
	exceptionID, err := service.Open(context.Background(), f.User, consignmentID, "delay")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Store.Exec(context.Background(), "UPDATE containers SET lease_owner='worker',lease_token='token' WHERE id=?", f.ContainerID); err != nil {
		t.Fatal(err)
	}
	f.Store.Hooks.BeforeAudit = func(context.Context) error { return errors.New("audit down") }
	if err := service.Resolve(context.Background(), f.User, exceptionID, "alternate", "failed-resolution"); err == nil {
		t.Fatal("resolution succeeded while its audit failed")
	}

	status, _, err := service.Get(context.Background(), f.User, exceptionID)
	if err != nil {
		t.Fatal(err)
	}
	var owner, token sql.NullString
	if err := f.Store.QueryRow(context.Background(), "SELECT lease_owner,lease_token FROM containers WHERE id=?", f.ContainerID).Scan(&owner, &token); err != nil {
		t.Fatal(err)
	}
	if status != "open" || !owner.Valid || owner.String != "worker" || !token.Valid || token.String != "token" {
		t.Fatalf("failed resolution leaked state: status=%s owner=%v token=%v", status, owner, token)
	}

	f.Store.Hooks.BeforeAudit = nil
	if err := service.Resolve(context.Background(), f.User, exceptionID, "alternate", "successful-resolution"); err != nil {
		t.Fatalf("resolution failed after audit recovery: %v", err)
	}
	status, _, err = service.Get(context.Background(), f.User, exceptionID)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Store.QueryRow(context.Background(), "SELECT lease_owner,lease_token FROM containers WHERE id=?", f.ContainerID).Scan(&owner, &token); err != nil {
		t.Fatal(err)
	}
	var events int
	if err := f.Store.QueryRow(context.Background(), "SELECT COUNT(*) FROM outbox_messages WHERE topic='exception.resolved' AND aggregate_id=?", exceptionID).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if status != "resolved" || owner.Valid || token.Valid || events != 1 {
		t.Fatalf("completed resolution status=%s owner=%v token=%v events=%d", status, owner, token, events)
	}
}
