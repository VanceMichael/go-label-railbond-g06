package train

import (
	"context"
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

type Plan struct {
	ID        string
	Number    string
	Capacity  int
	Departure time.Time
	SlotID    string
}

func (s Service) Create(ctx context.Context, u domain.User, corridorID, number string, capacity int, departure time.Time) (Plan, error) {
	if err := domain.RequireRole(u, "admin", "operator"); err != nil {
		return Plan{}, err
	}
	if capacity < 1 || number == "" {
		return Plan{}, fmt.Errorf("%w: train plan", domain.ErrInvalidState)
	}
	id, slot := storage.NewID(), storage.NewID()
	now := s.now().Format(time.RFC3339Nano)
	err := s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		if _, e := tx.Exec(ctx, "INSERT INTO rail_slots(id,tenant_id,corridor_id,starts_at,ends_at,status,version) VALUES(?,?,?,?,?,?,1)", slot, u.TenantID, corridorID, departure.Add(-2*time.Hour).Format(time.RFC3339Nano), departure.Add(2*time.Hour).Format(time.RFC3339Nano), "held"); e != nil {
			return e
		}
		if _, e := tx.Exec(ctx, "INSERT INTO trains(id,tenant_id,corridor_id,number,status,capacity,reserved,version,slot_id,departure_at,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)", id, u.TenantID, corridorID, number, string(domain.TrainPlanned), capacity, 0, 1, slot, departure.Format(time.RFC3339Nano), now); e != nil {
			return e
		}
		_, e := tx.Exec(ctx, "UPDATE rail_slots SET train_id=? WHERE id=?", id, slot)
		return e
	})
	return Plan{ID: id, Number: number, Capacity: capacity, Departure: departure, SlotID: slot}, err
}
func (s Service) Publish(ctx context.Context, u domain.User, id string) error {
	if err := domain.RequireRole(u, "admin", "operator"); err != nil {
		return err
	}
	res, err := s.Store.Exec(ctx, "UPDATE trains SET status=? WHERE tenant_id=? AND id=? AND status=?", domain.TrainPublished, u.TenantID, id, domain.TrainPlanned)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return fmt.Errorf("%w: publish train", domain.ErrInvalidState)
	}
	return nil
}
func (s Service) ReserveCapacity(ctx context.Context, u domain.User, id string, units int) error {
	if units < 1 {
		return domain.ErrInvalidState
	}
	if err := domain.CheckContext(ctx); err != nil {
		return err
	}
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var capacity, reserved int
		if err := tx.QueryRow(ctx, "SELECT capacity,reserved FROM trains WHERE tenant_id=? AND id=? AND status=?", u.TenantID, id, domain.TrainPublished).Scan(&capacity, &reserved); err != nil {
			return fmt.Errorf("%w: train capacity", domain.ErrNotFound)
		}
		if capacity-reserved < units {
			return fmt.Errorf("%w: train capacity", domain.ErrConflict)
		}
		res, err := tx.Exec(ctx, "UPDATE trains SET reserved=reserved+?,version=version+1 WHERE tenant_id=? AND id=? AND status=? AND capacity-reserved>=?", units, u.TenantID, id, domain.TrainPublished, units)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n != 1 {
			return fmt.Errorf("%w: capacity race", domain.ErrConflict)
		}
		return nil
	})
}
func (s Service) Depart(ctx context.Context, u domain.User, id string) error {
	if err := domain.RequireRole(u, "admin", "operator"); err != nil {
		return err
	}
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var status, slot string
		if err := tx.QueryRow(ctx, "SELECT status,slot_id FROM trains WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&status, &slot); err != nil {
			return err
		}
		if !domain.TrainStatus(status).CanMove(domain.TrainDeparted) {
			return fmt.Errorf("%w: train departure", domain.ErrInvalidState)
		}
		var holds int
		if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM customs_declarations d JOIN consignments c ON c.id=d.consignment_id WHERE d.tenant_id=? AND c.train_id=? AND d.status IN ('draft','submitted','hold')", u.TenantID, id).Scan(&holds); err != nil {
			return err
		}
		if holds > 0 {
			return fmt.Errorf("%w: customs not released", domain.ErrDeclarationHold)
		}
		if _, err := tx.Exec(ctx, "UPDATE trains SET status=?,version=version+1 WHERE tenant_id=? AND id=? AND status=?", domain.TrainDeparted, u.TenantID, id, domain.TrainPublished); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, "UPDATE rail_slots SET status='used' WHERE id=? AND status='held'", slot)
		return err
	})
}
func (s Service) Arrive(ctx context.Context, u domain.User, id string) error {
	res, err := s.Store.Exec(ctx, "UPDATE trains SET status=?,version=version+1 WHERE tenant_id=? AND id=? AND status=?", domain.TrainArrived, u.TenantID, id, domain.TrainDeparted)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return fmt.Errorf("%w: train arrival", domain.ErrInvalidState)
	}
	return nil
}
func (s Service) Get(ctx context.Context, u domain.User, id string) (storage.TrainRow, error) {
	return s.Store.GetTrain(ctx, u.TenantID, id)
}
func (s Service) ReleaseSlot(ctx context.Context, tenantID, slotID string) error {
	_, err := s.Store.Exec(ctx, "UPDATE rail_slots SET status='available',train_id=NULL,version=version+1 WHERE tenant_id=? AND id=? AND status IN ('held','used')", tenantID, slotID)
	return err
}
