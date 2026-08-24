package storage_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
)

func TestReadTxReleasesConnectionOnError(t *testing.T) {
	f := testkit.New(t)

	boom := errors.New("boom")
	if err := f.Store.WithReadTx(context.Background(), func(tx *sql.Tx) error { return boom }); err == nil {
		t.Fatal("expected callback error to propagate")
	}

	// After a failing read transaction the single connection must be released
	// so a subsequent short-deadline read does not time out.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := f.Store.WithReadTx(ctx, func(tx *sql.Tx) error { return nil }); err != nil {
		t.Fatalf("expected connection to be available after failed read tx, got: %v", err)
	}
}
