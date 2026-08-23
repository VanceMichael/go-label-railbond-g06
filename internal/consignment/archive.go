package consignment

import (
	"context"
	"fmt"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type ArchiveService struct{ Store *storage.Store }

func (s ArchiveService) Archive(ctx context.Context, u domain.User, id, requestID string) error {
	if err := domain.RequireRole(u, "admin", "operator"); err != nil {
		return err
	}
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var status string
		if err := tx.QueryRow(ctx, "SELECT status FROM consignments WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&status); err != nil {
			return err
		}
		if status != "delivered" {
			return fmt.Errorf("%w: archive requires delivery", domain.ErrInvalidState)
		}
		var invoices, docs int
		if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM invoices WHERE tenant_id=? AND consignment_id=? AND status='settled'", u.TenantID, id).Scan(&invoices); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM documents WHERE tenant_id=? AND consignment_id=? AND status='sealed'", u.TenantID, id).Scan(&docs); err != nil {
			return err
		}
		if invoices == 0 || docs == 0 {
			return fmt.Errorf("%w: archive prerequisites", domain.ErrInvalidState)
		}
		if _, err := tx.Exec(ctx, "UPDATE consignments SET status='archived',archived_at=?,version=version+1 WHERE tenant_id=? AND id=? AND status='delivered'", time.Now().UTC().Format(time.RFC3339Nano), u.TenantID, id); err != nil {
			return err
		}
		return s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "consignment.archived", "consignment", id, "success", requestID, "archive")
	})
}
