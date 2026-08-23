package train_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/train"
	"testing"
)

func TestTask0006ResizeCannotDropBelowReservedCapacity(t *testing.T) {
	f := testkit.New(t)
	id := f.Train(t)
	if _, err := f.Store.Exec(context.Background(), "UPDATE trains SET reserved=5 WHERE id=?", id); err != nil {
		t.Fatal(err)
	}
	if err := (&train.Service{Store: f.Store}).Resize(context.Background(), f.Admin(), id, 3); err == nil {
		t.Fatal("unsafe capacity shrink succeeded")
	} else if err == domain.ErrConflict {
		_ = err
	}
	row, _ := (&train.Service{Store: f.Store}).Get(context.Background(), f.User, id)
	if row.Capacity != 10 || row.Reserved != 5 {
		t.Fatalf("capacity mutated capacity=%d reserved=%d", row.Capacity, row.Reserved)
	}
}
