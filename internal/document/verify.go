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
)

type Verification struct {
	DocumentID, Hash string
	Items            int
	Valid            bool
}

func (s Service) Verify(ctx context.Context, u domain.User, id string) (Verification, error) {
	var v Verification
	var consignment, status, stored string
	if err := s.Store.QueryRow(ctx, "SELECT consignment_id,status,COALESCE(content_hash,'') FROM documents WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&consignment, &status, &stored); err != nil {
		return v, err
	}
	if status != "sealed" {
		return v, fmt.Errorf("%w: document not sealed", domain.ErrInvalidState)
	}
	rows, err := s.Store.Query(ctx, "SELECT sku,quantity,declared_value FROM consignment_items WHERE consignment_id=? ORDER BY id", consignment)
	if err != nil {
		return v, err
	}
	defer rows.Close()
	parts := []string{}
	for rows.Next() {
		var sku string
		var q, val int
		if err := rows.Scan(&sku, &q, &val); err != nil {
			return v, err
		}
		parts = append(parts, fmt.Sprintf("%s:%d:%d", sku, q, val))
	}
	sort.Strings(parts)
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	v = Verification{DocumentID: id, Hash: hex.EncodeToString(sum[:]), Items: len(parts), Valid: stored == hex.EncodeToString(sum[:])}
	if !v.Valid {
		return v, fmt.Errorf("%w: manifest hash mismatch", domain.ErrConflict)
	}
	return v, nil
}
func (s Service) RequireValid(ctx context.Context, u domain.User, id string) error {
	v, err := s.Verify(ctx, u, id)
	if err != nil {
		return err
	}
	if !v.Valid {
		return domain.ErrConflict
	}
	return nil
}

var _ = storage.NewID
