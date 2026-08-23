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
	returnHash := ""
	err := s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var consignment, status string
		var version int
		if err := tx.QueryRow(ctx, "SELECT consignment_id,status,version FROM documents WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&consignment, &status, &version); err != nil {
			return err
		}
		if status != "draft" {
			return fmt.Errorf("%w: document state", domain.ErrInvalidState)
		}
		rows, err := tx.Query(ctx, "SELECT sku,quantity,declared_value FROM consignment_items WHERE consignment_id=? ORDER BY id", consignment)
		if err != nil {
			return err
		}
		defer rows.Close()
		parts := []string{}
		for rows.Next() {
			var sku string
			var q, v int
			if err := rows.Scan(&sku, &q, &v); err != nil {
				return err
			}
			parts = append(parts, fmt.Sprintf("%s:%d:%d", sku, q, v))
		}
		sort.Strings(parts)
		sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
		returnHash = hex.EncodeToString(sum[:])
		if _, err := tx.Exec(ctx, "UPDATE documents SET status='sealed',content_hash=?,sealed_at=?,version=version+1 WHERE tenant_id=? AND id=? AND status='draft' AND version=?", returnHash, time.Now().UTC().Format(time.RFC3339Nano), u.TenantID, id, version); err != nil {
			return err
		}
		return s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "document.sealed", "document", id, "success", requestID, returnHash)
	})
	return returnHash, err
}
func (s Service) IsSealed(ctx context.Context, u domain.User, consignmentID string) (bool, error) {
	var n int
	err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM documents WHERE tenant_id=? AND consignment_id=? AND status='sealed'", u.TenantID, consignmentID).Scan(&n)
	return n > 0, err
}
