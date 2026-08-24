package dispatch

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type RebookService struct{ Store *storage.Store }

func (s RebookService) Rebook(ctx context.Context, u domain.User, consignmentID, carrier, key, requestID string) (string, error) {
	id := storage.NewID()
	err := s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		storedRoute, replay, err := s.Store.ObserveRebookReplay(ctx, tx, u.TenantID, key)
		if err != nil {
			return err
		}
		if replay {
			// A prior request with this idempotency key already completed and
			// produced a route assignment. Converge onto the first result and do
			// not create a second assignment, even if the client retried after a
			// timeout.
			id = storedRoute
			return nil
		}
		var status string
		if err := tx.QueryRow(ctx, "SELECT status FROM consignments WHERE tenant_id=? AND id=?", u.TenantID, consignmentID).Scan(&status); err != nil {
			return err
		}
		if status != "booked" {
			return fmt.Errorf("%w: rebook status", domain.ErrInvalidState)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO route_assignments(id,tenant_id,consignment_id,carrier,status,next_attempt_at) VALUES(?,?,?,?,?,?)", id, u.TenantID, consignmentID, carrier, "assigned", time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "INSERT INTO idempotency_keys(id,tenant_id,key,method,path,status_code,response_body,created_at) VALUES(?,?,?,?,?,?,?,?)", storage.NewID(), u.TenantID, key, "POST", "/rebook", 201, id, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			return err
		}
		return s.Store.RecordAudit(ctx, tx, u.TenantID, u.ID, "consignment.rebooked", "consignment", consignmentID, "success", requestID, carrier)
	})
	return id, err
}
