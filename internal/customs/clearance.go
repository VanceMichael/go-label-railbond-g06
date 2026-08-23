package customs

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type Clearance struct {
	DeclarationID, ConsignmentID, Status string
	ReleasedAt                           time.Time
}

func (s Service) Clear(ctx context.Context, u domain.User, id, requestID string) (Clearance, error) {
	var c Clearance
	err := s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var status, consignment string
		if err := tx.QueryRow(ctx, "SELECT status,consignment_id FROM customs_declarations WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&status, &consignment); err != nil {
			return err
		}
		if status != "submitted" {
			return fmt.Errorf("%w: clear from %s", domain.ErrInvalidState, status)
		}
		var holds int
		if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM customs_declarations WHERE tenant_id=? AND consignment_id=? AND status='hold'", u.TenantID, consignment).Scan(&holds); err != nil {
			return err
		}
		if holds > 0 {
			return domain.ErrDeclarationHold
		}
		c = Clearance{DeclarationID: id, ConsignmentID: consignment, Status: "released", ReleasedAt: time.Now().UTC()}
		if _, err := tx.Exec(ctx, "UPDATE customs_declarations SET status='released',released_at=?,version=version+1 WHERE tenant_id=? AND id=? AND status='submitted'", c.ReleasedAt.Format(time.RFC3339Nano), u.TenantID, id); err != nil {
			return err
		}
		if err := s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "customs.cleared", "declaration", id, "success", requestID, consignment); err != nil {
			return err
		}
		return s.Store.Enqueue(ctx, tx, u.TenantID, "customs.cleared", id, consignment)
	})
	return c, err
}
func (s Service) Reopen(ctx context.Context, u domain.User, id, reason string) error {
	if reason == "" {
		return fmt.Errorf("%w: reopen reason", domain.ErrInvalidState)
	}
	res, err := s.Store.Exec(ctx, "UPDATE customs_declarations SET status='hold',hold_reason=?,version=version+1 WHERE tenant_id=? AND id=? AND status='released'", reason, u.TenantID, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}
func (s Service) ReleasedAt(ctx context.Context, u domain.User, id string) (time.Time, error) {
	var raw string
	if err := s.Store.QueryRow(ctx, "SELECT released_at FROM customs_declarations WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&raw); err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, raw)
}
