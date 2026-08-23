package document

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"sort"
	"strings"
	"time"
)

type Service struct{ Store *storage.Store }

func (s Service) Create(ctx context.Context, u domain.User, consignmentID, kind string) (string, error) {
	id := storage.NewID()
	_, err := s.Store.Exec(ctx, "INSERT INTO documents(id,tenant_id,consignment_id,kind,status,created_at) VALUES(?,?,?,?,?,?)", id, u.TenantID, consignmentID, kind, "draft", time.Now().UTC().Format(time.RFC3339Nano))
	return id, err
}
func (s Service) SealManifest(ctx context.Context, u domain.User, id string, requestID string) (string, error) {
	var consignment, status string
	var version int
	if err := s.Store.QueryRow(ctx, "SELECT consignment_id,status,version FROM documents WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&consignment, &status, &version); err != nil {
		return "", err
	}
	if status != "draft" {
		return "", fmt.Errorf("%w: document state", domain.ErrInvalidState)
	}
	rows, err := s.Store.Query(ctx, "SELECT sku,quantity,declared_value FROM consignment_items WHERE consignment_id=? ORDER BY id", consignment)
	if err != nil {
		return "", err
	}
	parts := []string{}
	for rows.Next() {
		var sku string
		var quantity, value int
		if err := rows.Scan(&sku, &quantity, &value); err != nil {
			_ = rows.Close()
			return "", err
		}
		parts = append(parts, fmt.Sprintf("%s:%d:%d", sku, quantity, value))
	}
	if err := rows.Close(); err != nil {
		return "", err
	}
	sort.Strings(parts)
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	contentHash := hex.EncodeToString(sum[:])
	if err := s.Store.SealDocument(ctx, u.TenantID, id, version, contentHash, time.Now().UTC()); err != nil {
		return "", err
	}
	err = s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		return s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "document.sealed", "document", id, "success", requestID, contentHash)
	})
	return contentHash, err
}
func (s Service) IsSealed(ctx context.Context, u domain.User, consignmentID string) (bool, error) {
	var n int
	err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM documents WHERE tenant_id=? AND consignment_id=? AND status='sealed'", u.TenantID, consignmentID).Scan(&n)
	return n > 0, err
}
