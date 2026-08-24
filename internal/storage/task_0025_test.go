package storage_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
	"time"
)

func TestTask0025ReadTransactionReleasesConnectionOnError(t *testing.T) {
	f := testkit.New(t)
	f.Store.DB.SetMaxOpenConns(1)
	if err := f.Store.WithReadTx(context.Background(), func(*sql.Tx) error { return errors.New("read failed") }); err == nil {
		t.Fatal("expected read failure")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := f.Store.Ping(ctx); err != nil {
		t.Fatalf("connection remained held: %v", err)
	}
}
