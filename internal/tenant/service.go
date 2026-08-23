package tenant

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type Service struct{ Store *storage.Store }

func (s Service) CreateCorridor(ctx context.Context, u domain.User, name, origin, destination, tz string) (string, error) {
	if err := domain.RequireRole(u, "admin", "operator"); err != nil {
		return "", err
	}
	id := storage.NewID()
	_, err := s.Store.Exec(ctx, "INSERT INTO corridors(id,tenant_id,name,origin,destination,timezone,created_at) VALUES(?,?,?,?,?,?,?)", id, u.TenantID, name, origin, destination, tz, time.Now().UTC().Format(time.RFC3339Nano))
	return id, err
}
func (s Service) AddWarehouseSlot(ctx context.Context, u domain.User, code, zone string) (string, error) {
	if err := domain.RequireRole(u, "admin", "operator"); err != nil {
		return "", err
	}
	id := storage.NewID()
	_, err := s.Store.Exec(ctx, "INSERT INTO warehouse_slots(id,tenant_id,code,zone,status,version) VALUES(?,?,?,?,?,1)", id, u.TenantID, code, zone, "available")
	return id, err
}
func (s Service) AddCheckpoint(ctx context.Context, u domain.User, corridorID string, seq int, name string, inspection bool) (string, error) {
	if err := domain.RequireRole(u, "admin", "operator"); err != nil {
		return "", err
	}
	id := storage.NewID()
	_, err := s.Store.Exec(ctx, "INSERT INTO checkpoints(id,tenant_id,corridor_id,sequence_no,name,required_inspection) VALUES(?,?,?,?,?,?)", id, u.TenantID, corridorID, seq, name, boolInt(inspection))
	return id, err
}
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
func IsNotFound(err error) bool { return err == sql.ErrNoRows }
func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}
