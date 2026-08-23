package document_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/document"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
)

func TestSealManifestAndReadSealed(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	s := document.Service{Store: f.Store}
	id, err := s.Create(context.Background(), f.User, c, "manifest")
	if err != nil {
		t.Fatal(err)
	}
	hash, err := s.SealManifest(context.Background(), f.User, id, "r")
	if err != nil || len(hash) != 64 {
		t.Fatalf("%v %s", err, hash)
	}
	ok, err := s.IsSealed(context.Background(), f.User, c)
	if err != nil || !ok {
		t.Fatalf("%v %v", err, ok)
	}
}
func TestSecondSealRejected(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	s := document.Service{Store: f.Store}
	id, _ := s.Create(context.Background(), f.User, c, "manifest")
	_, _ = s.SealManifest(context.Background(), f.User, id, "r")
	if _, err := s.SealManifest(context.Background(), f.User, id, "r"); err == nil {
		t.Fatal("sealed document resealed")
	}
}
