package storage_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestTransactionRollsBackOnError(t *testing.T) {
	f := testkit.New(t)
	err := f.Store.WithTx(context.Background(), func(tx *storage.Tx) error { return errors.New("stop") })
	if err == nil {
		t.Fatal("expected rollback error")
	}
}
func TestAuditHookErrorIsWrapped(t *testing.T) {
	f := testkit.New(t)
	f.Store.Hooks.BeforeAudit = func(context.Context) error { return domain.ErrConflict }
	err := f.Store.WithTx(context.Background(), func(tx *storage.Tx) error {
		return f.Store.RecordAudit(context.Background(), tx, f.TenantID, f.User.ID, "x", "x", "x", "failed", "r", "d")
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatal(err)
	}
}
