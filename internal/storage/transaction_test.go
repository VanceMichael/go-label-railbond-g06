package storage_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestPingForeignKeysAndOptimize(t *testing.T) {
	f := testkit.New(t)
	if err := f.Store.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	ok, err := f.Store.ForeignKeys(context.Background())
	if err != nil || !ok {
		t.Fatalf("%v %v", err, ok)
	}
	if err := f.Store.Vacuum(context.Background()); err != nil {
		t.Fatal(err)
	}
}
