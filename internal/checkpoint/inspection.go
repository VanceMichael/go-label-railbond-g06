package checkpoint

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"time"
)

type InspectionRecord struct {
	ID, Status, Result string
	CompletedAt        sql.NullString
}

func (s Service) Inspection(ctx context.Context, u domain.User, consignmentID, checkpointID string) (InspectionRecord, error) {
	var r InspectionRecord
	err := s.Store.QueryRow(ctx, "SELECT id,status,COALESCE(result,''),COALESCE(completed_at,'') FROM inspections WHERE tenant_id=? AND consignment_id=? AND checkpoint_id=?", u.TenantID, consignmentID, checkpointID).Scan(&r.ID, &r.Status, &r.Result, &r.CompletedAt)
	if err == sql.ErrNoRows {
		return r, domain.ErrNotFound
	}
	return r, err
}
func (s Service) FailInspection(ctx context.Context, u domain.User, consignmentID, checkpointID, reason string) error {
	res, err := s.Store.Exec(ctx, "UPDATE inspections SET status='failed',result=?,completed_at=? WHERE tenant_id=? AND consignment_id=? AND checkpoint_id=? AND status='pending'", reason, time.Now().UTC().Format(time.RFC3339Nano), u.TenantID, consignmentID, checkpointID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return fmt.Errorf("%w: inspection transition", domain.ErrInvalidState)
	}
	return nil
}
func (s Service) ResetInspection(ctx context.Context, u domain.User, consignmentID, checkpointID string) error {
	_, err := s.Store.Exec(ctx, "UPDATE inspections SET status='pending',result=NULL,completed_at=NULL WHERE tenant_id=? AND consignment_id=? AND checkpoint_id=? AND status='failed'", u.TenantID, consignmentID, checkpointID)
	return err
}
