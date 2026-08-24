package document_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/document"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestVerifySealedManifest(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	s := document.Service{Store: f.Store}
	id, _ := s.Create(context.Background(), f.User, c, "manifest")
	if _, err := s.SealManifest(context.Background(), f.User, id, "r"); err != nil {
		t.Fatal(err)
	}
	v, err := s.Verify(context.Background(), f.User, id)
	if err != nil || !v.Valid || v.Items != 1 {
		t.Fatalf("%v %#v", err, v)
	}
}

func TestVerifyAbortsOnCancelledContext(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	s := document.Service{Store: f.Store}
	id, _ := s.Create(context.Background(), f.User, c, "manifest")
	if _, err := s.SealManifest(context.Background(), f.User, id, "r"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.Verify(ctx, f.User, id); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
