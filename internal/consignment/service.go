package consignment

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type Service struct {
	Store *storage.Store
	Now   func() time.Time
}

func (s Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now().UTC()
}

type Item struct {
	SKU, Description        string
	Quantity, DeclaredValue int
}
type CreateInput struct {
	TrainID, ContainerID, Reference string
	Items                           []Item
}
type Record struct {
	ID, Reference, Status string
	TrainID, ContainerID  string
}

func (s Service) Create(ctx context.Context, u domain.User, in CreateInput, requestID string) (Record, error) {
	if err := domain.RequireRole(u, "admin", "operator"); err != nil {
		return Record{}, err
	}
	if len(in.Items) == 0 || in.Reference == "" {
		return Record{}, fmt.Errorf("%w: consignment items", domain.ErrInvalidState)
	}
	id := storage.NewID()
	now := s.now().Format(time.RFC3339Nano)
	err := s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var trainStatus, containerStatus string
		if err := tx.QueryRow(ctx, "SELECT status FROM trains WHERE tenant_id=? AND id=?", u.TenantID, in.TrainID).Scan(&trainStatus); err != nil {
			return fmt.Errorf("%w: train", domain.ErrNotFound)
		}
		if trainStatus != "published" {
			return fmt.Errorf("%w: train not published", domain.ErrInvalidState)
		}
		if err := tx.QueryRow(ctx, "SELECT status FROM containers WHERE tenant_id=? AND id=?", u.TenantID, in.ContainerID).Scan(&containerStatus); err != nil {
			return fmt.Errorf("%w: container", domain.ErrNotFound)
		}
		if containerStatus != "available" {
			return fmt.Errorf("%w: container unavailable", domain.ErrConflict)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO consignments(id,tenant_id,train_id,container_id,reference,status,created_at) VALUES(?,?,?,?,?,?,?)", id, u.TenantID, in.TrainID, in.ContainerID, in.Reference, string(domain.ConsignmentDraft), now); err != nil {
			return err
		}
		for _, item := range in.Items {
			if item.Quantity <= 0 {
				return fmt.Errorf("%w: item quantity", domain.ErrInvalidState)
			}
			if _, err := tx.Exec(ctx, "INSERT INTO consignment_items(id,consignment_id,sku,description,quantity,declared_value) VALUES(?,?,?,?,?,?)", storage.NewID(), id, item.SKU, item.Description, item.Quantity, item.DeclaredValue); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, "UPDATE consignments SET status='booked',version=version+1 WHERE id=? AND status='draft'", id); err != nil {
			return err
		}
		return s.Store.Enqueue(ctx, tx, u.TenantID, "consignment.created", id, in.Reference)
	})
	if err == nil {
		if auditErr := s.recordCreateAuditAfterCommit(u, id, requestID, in.Reference); auditErr != nil {
			return Record{}, domain.Wrap("create consignment audit", auditErr)
		}
	}
	if err != nil {
		return Record{}, domain.Wrap("create consignment", err)
	}
	return Record{ID: id, Reference: in.Reference, Status: string(domain.ConsignmentBooked), TrainID: in.TrainID, ContainerID: in.ContainerID}, nil
}

func (s Service) Get(ctx context.Context, u domain.User, id string) (Record, error) {
	r, err := s.Store.GetConsignment(ctx, u.TenantID, id)
	if err == sql.ErrNoRows {
		return Record{}, domain.ErrNotFound
	}
	if err != nil {
		return Record{}, err
	}
	return Record{ID: r.ID, Reference: r.Reference, Status: r.Status, TrainID: r.TrainID, ContainerID: r.ContainerID}, nil
}
func (s Service) Advance(ctx context.Context, u domain.User, id string, to domain.ConsignmentStatus) error {
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var status string
		var version int
		if err := tx.QueryRow(ctx, "SELECT status,version FROM consignments WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&status, &version); err != nil {
			return err
		}
		if !domain.ConsignmentStatus(status).CanMove(to) {
			return fmt.Errorf("%w: consignment transition", domain.ErrInvalidState)
		}
		res, err := tx.Exec(ctx, "UPDATE consignments SET status=?,version=version+1 WHERE tenant_id=? AND id=? AND status=? AND version=?", to, u.TenantID, id, status, version)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n != 1 {
			return domain.ErrConflict
		}
		return nil
	})
}
