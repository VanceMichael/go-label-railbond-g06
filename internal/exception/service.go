package exception

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type Service struct{ Store *storage.Store }

func (s Service) Open(ctx context.Context, u domain.User, consignmentID, reason string) (string, error) {
	id := storage.NewID()
	_, err := s.Store.Exec(ctx, "INSERT INTO exceptions(id,tenant_id,consignment_id,status,reason,opened_at) VALUES(?,?,?,?,?,?)", id, u.TenantID, consignmentID, "open", reason, time.Now().UTC().Format(time.RFC3339Nano))
	return id, err
}
func (s Service) Resolve(ctx context.Context, u domain.User, id, replacement, requestID string) error {
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var status, consignment string
		if err := tx.QueryRow(ctx, "SELECT status,consignment_id FROM exceptions WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&status, &consignment); err != nil {
			return err
		}
		if status != "open" {
			return fmt.Errorf("%w: exception state", domain.ErrInvalidState)
		}
		if _, err := tx.Exec(ctx, "UPDATE exceptions SET status='resolved',replacement_route=?,resolved_at=? WHERE tenant_id=? AND id=? AND status='open'", replacement, time.Now().UTC().Format(time.RFC3339Nano), u.TenantID, id); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "UPDATE containers SET lease_owner=NULL,lease_token=NULL,lease_until=NULL WHERE tenant_id=? AND id=(SELECT container_id FROM consignments WHERE id=?)", u.TenantID, consignment); err != nil {
			return err
		}
		if err := s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "exception.resolved", "exception", id, "success", requestID, replacement); err != nil {
			return err
		}
		return s.Store.Enqueue(ctx, tx, u.TenantID, "exception.resolved", id, consignment)
	})
}
func (s Service) Get(ctx context.Context, u domain.User, id string) (string, string, error) {
	var st, reason string
	err := s.Store.QueryRow(ctx, "SELECT status,reason FROM exceptions WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&st, &reason)
	return st, reason, err
}
