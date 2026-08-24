package document_test

import (
	"context"
	"errors"
	"testing"

	"github.com/VanceMichael/go-base-railbond-g06/internal/document"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
)

func TestTask0018CancelledManifestVerificationStopsDatabaseRead(t *testing.T) {
	f := testkit.New(t)
	consignmentID := f.Consignment(t, f.Train(t))
	service := document.Service{Store: f.Store}
	documentID, err := service.Create(context.Background(), f.User, consignmentID, "bonded-manifest")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SealManifest(context.Background(), f.User, documentID, "seal-request"); err != nil {
		t.Fatal(err)
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.Verify(cancelled, f.User, documentID); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled verification error = %v, want context canceled", err)
	}

	verification, err := service.Verify(context.Background(), f.User, documentID)
	if err != nil || !verification.Valid || verification.Items != 1 {
		t.Fatalf("normal verification=%#v err=%v", verification, err)
	}
}
