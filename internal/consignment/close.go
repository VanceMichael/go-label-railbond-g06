package consignment

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type CloseReport struct {
	ConsignmentID                string
	ReleasedSlots, OpenDocuments int
	ClosedAt                     time.Time
}

func (s Service) Close(ctx context.Context, u domain.User, id, requestID string) (CloseReport, error) {
	if err := domain.RequireRole(u, "admin", "operator"); err != nil {
		return CloseReport{}, err
	}
	report := CloseReport{ConsignmentID: id, ClosedAt: time.Now().UTC()}
	persistenceContext := archivePersistenceContext(ctx)
	err := s.Store.WithTx(persistenceContext, func(tx *storage.Tx) error {
		ctx = persistenceContext
		var status string
		if err := tx.QueryRow(ctx, "SELECT status FROM consignments WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&status); err != nil {
			return err
		}
		if status != "delivered" {
			return fmt.Errorf("%w: close requires delivery", domain.ErrInvalidState)
		}
		var open int
		if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM documents WHERE tenant_id=? AND consignment_id=? AND status!='sealed'", u.TenantID, id).Scan(&open); err != nil {
			return err
		}
		if open > 0 {
			return fmt.Errorf("%w: open documents", domain.ErrInvalidState)
		}
		var active int
		if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM slot_reservations WHERE tenant_id=? AND consignment_id=? AND status='active'", u.TenantID, id).Scan(&active); err != nil {
			return err
		}
		if active > 0 {
			return fmt.Errorf("%w: active warehouse reservation", domain.ErrInvalidState)
		}
		if _, err := tx.Exec(ctx, "UPDATE consignments SET status='archived',archived_at=?,version=version+1 WHERE tenant_id=? AND id=? AND status='delivered'", report.ClosedAt.Format(time.RFC3339Nano), u.TenantID, id); err != nil {
			return err
		}
		report.ReleasedSlots = active
		report.OpenDocuments = open
		return s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "consignment.closed", "consignment", id, "success", requestID, "archive prerequisites satisfied")
	})
	return report, err
}
func (s Service) CanClose(ctx context.Context, u domain.User, id string) (bool, error) {
	var status string
	if err := s.Store.QueryRow(ctx, "SELECT status FROM consignments WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&status); err != nil {
		return false, err
	}
	if status != "delivered" {
		return false, nil
	}
	var open int
	if err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM documents WHERE tenant_id=? AND consignment_id=? AND status!='sealed'", u.TenantID, id).Scan(&open); err != nil {
		return false, err
	}
	return open == 0, nil
}
func (s Service) ArchiveDue(ctx context.Context, u domain.User, before time.Time) ([]string, error) {
	rows, err := s.Store.Query(ctx, "SELECT id FROM consignments WHERE tenant_id=? AND status='delivered' AND delivered_at<? ORDER BY delivered_at", u.TenantID, before.Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
