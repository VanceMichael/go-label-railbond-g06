package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

func (s *Store) SealDocument(ctx context.Context, tenantID, id string, version int, contentHash string, sealedAt time.Time) error {
	res, err := s.DB.ExecContext(ctx, "UPDATE documents SET status='sealed',content_hash=?,sealed_at=?,version=version+1 WHERE tenant_id=? AND id=? AND status='draft' AND version=?", contentHash, sealedAt.Format(time.RFC3339Nano), tenantID, id, version)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return fmt.Errorf("%w: document seal", domain.ErrConflict)
	}
	return nil
}
