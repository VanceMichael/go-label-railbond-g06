package checkpoint

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type Service struct{ Store *storage.Store }
type Scan struct {
	ID           string
	Sequence     int
	EvidenceHash string
}

func (s Service) RecordScan(ctx context.Context, u domain.User, consignmentID, checkpointID, scanner, evidence, requestID string) (Scan, error) {
	if err := domain.CheckContext(ctx); err != nil {
		return Scan{}, err
	}
	if s.Store.Hooks.BeforeCheckpoint != nil {
		if err := s.Store.Hooks.BeforeCheckpoint(ctx); err != nil {
			return Scan{}, err
		}
	}
	var result Scan
	err := s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var seq, required, progress int
		var cStatus string
		if err := tx.QueryRow(ctx, "SELECT sequence_no,required_inspection FROM checkpoints WHERE tenant_id=? AND id=?", u.TenantID, checkpointID).Scan(&seq, &required); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, "SELECT status,current_checkpoint FROM consignments WHERE tenant_id=? AND id=?", u.TenantID, consignmentID).Scan(&cStatus, &progress); err != nil {
			return err
		}
		if cStatus != "in_transit" && cStatus != "at_checkpoint" {
			return fmt.Errorf("%w: scan state", domain.ErrInvalidState)
		}
		var exists int
		if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM checkpoint_events WHERE tenant_id=? AND consignment_id=? AND checkpoint_id=?", u.TenantID, consignmentID, checkpointID).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			return fmt.Errorf("%w: duplicate checkpoint", domain.ErrConflict)
		}
		if required == 1 {
			var inspection string
			if err := tx.QueryRow(ctx, "SELECT status FROM inspections WHERE tenant_id=? AND consignment_id=? AND checkpoint_id=?", u.TenantID, consignmentID, checkpointID).Scan(&inspection); err != nil {
				return err
			}
			if inspection != "passed" {
				return fmt.Errorf("%w: inspection pending", domain.ErrInvalidState)
			}
		}
		id := storage.NewID()
		now := time.Now().UTC().Format(time.RFC3339Nano)
		if _, err := tx.Exec(ctx, "INSERT INTO checkpoint_events(id,tenant_id,checkpoint_id,consignment_id,scanner_id,evidence_hash,observed_at,created_at) VALUES(?,?,?,?,?,?,?,?)", id, u.TenantID, checkpointID, consignmentID, scanner, evidence, now, now); err != nil {
			return err
		}
		if err := advanceCheckpointProgress(ctx, tx, u.TenantID, consignmentID, seq); err != nil {
			return err
		}
		if err := s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "checkpoint.scanned", "consignment", consignmentID, "success", requestID, evidence); err != nil {
			return err
		}
		result = Scan{ID: id, Sequence: seq, EvidenceHash: evidence}
		return nil
	})
	if err != nil {
		return Scan{}, domain.Wrap("record checkpoint", err)
	}
	return result, nil
}
func (s Service) CompleteCheckpoint(ctx context.Context, u domain.User, consignmentID, checkpointID, requestID string) error {
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var seq, required int
		var status string
		if err := tx.QueryRow(ctx, "SELECT sequence_no,required_inspection FROM checkpoints WHERE tenant_id=? AND id=?", u.TenantID, checkpointID).Scan(&seq, &required); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, "SELECT status FROM consignments WHERE tenant_id=? AND id=?", u.TenantID, consignmentID).Scan(&status); err != nil {
			return err
		}
		if status != "at_checkpoint" {
			return fmt.Errorf("%w: checkpoint completion", domain.ErrInvalidState)
		}
		if required == 1 {
			var inspection string
			if err := tx.QueryRow(ctx, "SELECT status FROM inspections WHERE tenant_id=? AND consignment_id=? AND checkpoint_id=?", u.TenantID, consignmentID, checkpointID).Scan(&inspection); err != nil {
				return err
			}
			if inspection != "passed" {
				return fmt.Errorf("%w: inspection pending", domain.ErrInvalidState)
			}
		}
		if _, err := tx.Exec(ctx, "UPDATE consignments SET status='in_transit',version=version+1 WHERE tenant_id=? AND id=? AND status='at_checkpoint'", u.TenantID, consignmentID); err != nil {
			return err
		}
		return s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "checkpoint.completed", "consignment", consignmentID, "success", requestID, fmt.Sprintf("sequence=%d", seq))
	})
}
func (s Service) CreateInspection(ctx context.Context, u domain.User, consignmentID, checkpointID string) error {
	_, err := s.Store.Exec(ctx, "INSERT INTO inspections(id,tenant_id,checkpoint_id,consignment_id,status) VALUES(?,?,?,?,?)", storage.NewID(), u.TenantID, checkpointID, consignmentID, "pending")
	return err
}
func (s Service) PassInspection(ctx context.Context, u domain.User, consignmentID, checkpointID string) error {
	_, err := s.Store.Exec(ctx, "UPDATE inspections SET status='passed',result='clear',completed_at=? WHERE tenant_id=? AND consignment_id=? AND checkpoint_id=? AND status='pending'", time.Now().UTC().Format(time.RFC3339Nano), u.TenantID, consignmentID, checkpointID)
	return err
}
