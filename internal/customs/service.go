package customs

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type Service struct{ Store *storage.Store }
type Declaration struct {
	ID, ConsignmentID, Status string
	HoldReason                string
}

func (s Service) Create(ctx context.Context, u domain.User, consignmentID string) (Declaration, error) {
	id := storage.NewID()
	_, err := s.Store.Exec(ctx, "INSERT INTO customs_declarations(id,tenant_id,consignment_id,status,version) VALUES(?,?,?,?,1)", id, u.TenantID, consignmentID, domain.DeclarationDraft)
	if err != nil {
		return Declaration{}, err
	}
	return Declaration{ID: id, ConsignmentID: consignmentID, Status: string(domain.DeclarationDraft)}, nil
}
func (s Service) Submit(ctx context.Context, u domain.User, id, requestID string) error {
	if err := domain.CheckContext(ctx); err != nil {
		return err
	}
	var consignment string
	err := s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var status string
		if err := tx.QueryRow(ctx, "SELECT status,consignment_id FROM customs_declarations WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&status, &consignment); err != nil {
			return err
		}
		if !domain.DeclarationStatus(status).CanMove(domain.DeclarationSubmitted) {
			return fmt.Errorf("%w: declaration submit", domain.ErrInvalidState)
		}
		var itemCount int
		if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM consignment_items WHERE consignment_id=?", consignment).Scan(&itemCount); err != nil {
			return err
		}
		if itemCount == 0 {
			return fmt.Errorf("%w: declaration items", domain.ErrInvalidState)
		}
		if _, err := tx.Exec(ctx, "UPDATE customs_declarations SET status='submitted',submitted_at=?,version=version+1 WHERE tenant_id=? AND id=? AND status=?", time.Now().UTC().Format(time.RFC3339Nano), u.TenantID, id, domain.DeclarationDraft); err != nil {
			return err
		}
		return s.Store.Enqueue(ctx, tx, u.TenantID, "customs.submitted", id, consignment)
	})
	if err != nil {
		return err
	}
	return s.recordSubmitAuditAfterCommit(u, id, requestID, consignment)
}
func (s Service) Release(ctx context.Context, u domain.User, id, requestID string) error {
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var status string
		var consignment string
		if err := tx.QueryRow(ctx, "SELECT status,consignment_id FROM customs_declarations WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&status, &consignment); err != nil {
			return err
		}
		if status == "hold" {
			return fmt.Errorf("%w: declaration %s", domain.ErrDeclarationHold, id)
		}
		if status != "submitted" {
			return fmt.Errorf("%w: release from %s", domain.ErrInvalidState, status)
		}
		if _, err := tx.Exec(ctx, "UPDATE customs_declarations SET status='released',released_at=?,version=version+1 WHERE tenant_id=? AND id=? AND status='submitted'", time.Now().UTC().Format(time.RFC3339Nano), u.TenantID, id); err != nil {
			return err
		}
		if err := s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "customs.released", "declaration", id, "success", requestID, consignment); err != nil {
			return err
		}
		return s.Store.Enqueue(ctx, tx, u.TenantID, "customs.released", id, consignment)
	})
}
func (s Service) Hold(ctx context.Context, u domain.User, id, reason string) error {
	_, err := s.Store.Exec(ctx, "UPDATE customs_declarations SET status='hold',hold_reason=?,version=version+1 WHERE tenant_id=? AND id=? AND status IN ('submitted','hold')", reason, u.TenantID, id)
	return err
}
func (s Service) Get(ctx context.Context, u domain.User, id string) (Declaration, error) {
	r, err := s.Store.GetDeclaration(ctx, u.TenantID, id)
	if err == sql.ErrNoRows {
		return Declaration{}, domain.ErrNotFound
	}
	return Declaration{ID: r.ID, ConsignmentID: r.ConsignmentID, Status: r.Status, HoldReason: r.HoldReason.String}, err
}
