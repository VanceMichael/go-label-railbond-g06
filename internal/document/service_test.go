package document_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/document"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
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

func TestSealRollsBackWhenAuditFails(t *testing.T) {
	f := testkit.New(t)
	c := f.Consignment(t, f.Train(t))
	s := document.Service{Store: f.Store}
	id, _ := s.Create(context.Background(), f.User, c, "manifest")

	f.Store.Hooks.BeforeAudit = func(context.Context) error { return domain.ErrConflict }
	if _, err := s.SealManifest(context.Background(), f.User, id, "r"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	f.Store.Hooks.BeforeAudit = nil

	var status string
	var hash *string
	if err := f.Store.QueryRow(context.Background(), "SELECT status,content_hash FROM documents WHERE tenant_id=? AND id=?", f.TenantID, id).Scan(&status, &hash); err != nil {
		t.Fatal(err)
	}
	if status != "draft" || hash != nil {
		t.Fatalf("seal not rolled back: status=%s hash=%v", status, hash)
	}

	hash2, err := s.SealManifest(context.Background(), f.User, id, "r")
	if err != nil || len(hash2) != 64 {
		t.Fatalf("retry failed: %v %s", err, hash2)
	}
}
