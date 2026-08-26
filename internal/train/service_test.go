package train_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/train"
	"sync"
	"testing"
	"time"
)

func TestCreatePublishesAndLists(t *testing.T) {
	f := testkit.New(t)
	s := train.Service{Store: f.Store}
	p, err := s.Create(context.Background(), f.Admin(), f.CorridorID, "RB100", 4, time.Now().UTC().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Publish(context.Background(), f.Admin(), p.ID); err != nil {
		t.Fatal(err)
	}
	items, err := s.List(context.Background(), f.User)
	if err != nil || len(items) != 1 {
		t.Fatalf("%v %d", err, len(items))
	}
}
func TestConcurrentReservationsRespectCapacity(t *testing.T) {
	f := testkit.New(t)
	s := train.Service{Store: f.Store}
	id := f.Train(t)
	var wg sync.WaitGroup
	var mu sync.Mutex
	success := 0
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if s.ReserveCapacity(context.Background(), f.User, id, 3) == nil {
				mu.Lock()
				success++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if success > 3 {
		t.Fatalf("oversold with %d successes", success)
	}
	row, err := s.Get(context.Background(), f.User, id)
	if err != nil {
		t.Fatal(err)
	}
	if row.Reserved > row.Capacity {
		t.Fatal("stored oversold capacity")
	}
}
func TestDepartRejectsCustomsHold(t *testing.T) {
	f := testkit.New(t)
	s := train.Service{Store: f.Store}
	id := f.Train(t)
	c := f.Consignment(t, id)
	d := f.Declaration(t, c, "hold")
	_ = d
	if err := s.Depart(context.Background(), f.User, id); !errors.Is(err, domain.ErrDeclarationHold) {
		t.Fatalf("want hold, got %v", err)
	}
	row, _ := s.Get(context.Background(), f.User, id)
	if row.Status != "published" {
		t.Fatal("train departed")
	}
}
func TestCancelReleasesSlot(t *testing.T) {
	f := testkit.New(t)
	s := train.Service{Store: f.Store}
	id := f.Train(t)
	if err := (train.CancellationService{Store: f.Store}).Cancel(context.Background(), f.Admin(), id); err != nil {
		t.Fatal(err)
	}
	row, _ := s.Get(context.Background(), f.User, id)
	if row.Status != "cancelled" {
		t.Fatal(row.Status)
	}
}
