package worker_test

import (
	"github.com/VanceMichael/go-base-railbond-g06/internal/worker"
	"testing"
	"time"
)

func TestBackoffCaps(t *testing.T) {
	b := worker.Backoff{Base: time.Second, Max: time.Minute}
	if b.Delay(0) != time.Second || b.Delay(20) != time.Minute {
		t.Fatal(b.Delay(0), b.Delay(20))
	}
	if !worker.Permanent(3, 3) {
		t.Fatal("not permanent")
	}
}
