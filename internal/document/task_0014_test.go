package document_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/VanceMichael/go-base-railbond-g06/internal/document"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
)

func TestTask0014SealAuditFailureLeavesManifestDraft(t *testing.T) {
	f := testkit.New(t)
	consignmentID := f.Consignment(t, f.Train(t))
	service := document.Service{Store: f.Store}
	documentID, err := service.Create(context.Background(), f.User, consignmentID, "bonded-manifest")
	if err != nil {
		t.Fatal(err)
	}
	f.Store.Hooks.BeforeAudit = func(context.Context) error { return errors.New("audit unavailable") }

	if _, err := service.SealManifest(context.Background(), f.User, documentID, "request-failed"); err == nil {
		t.Fatal("manifest seal succeeded while audit persistence failed")
	}
	var status string
	var contentHash sql.NullString
	if err := f.Store.QueryRow(context.Background(), "SELECT status,content_hash FROM documents WHERE id=?", documentID).Scan(&status, &contentHash); err != nil {
		t.Fatal(err)
	}
	if status != "draft" || contentHash.Valid {
		t.Fatalf("failed seal leaked durable state: status=%s hash=%v", status, contentHash)
	}

	f.Store.Hooks.BeforeAudit = nil
	hash, err := service.SealManifest(context.Background(), f.User, documentID, "request-success")
	if err != nil {
		t.Fatalf("manifest could not be sealed after audit recovery: %v", err)
	}
	verification, err := service.Verify(context.Background(), f.User, documentID)
	if err != nil || !verification.Valid || verification.Hash != hash {
		t.Fatalf("sealed manifest verification=%#v err=%v", verification, err)
	}
}
