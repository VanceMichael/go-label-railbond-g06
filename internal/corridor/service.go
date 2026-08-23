package corridor

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type Service struct{ Store *storage.Store }

func (s Service) Get(ctx context.Context, u domain.User, id string) (string, string, string, error) {
	var origin, destination, tz string
	err := s.Store.QueryRow(ctx, "SELECT origin,destination,timezone FROM corridors WHERE tenant_id=? AND id=? AND active=1", u.TenantID, id).Scan(&origin, &destination, &tz)
	if err == sql.ErrNoRows {
		return "", "", "", fmt.Errorf("%w: corridor", domain.ErrNotFound)
	}
	return origin, destination, tz, err
}
func (s Service) Deactivate(ctx context.Context, u domain.User, id string) error {
	if err := domain.RequireRole(u, "admin"); err != nil {
		return err
	}
	_, err := s.Store.Exec(ctx, "UPDATE corridors SET active=0 WHERE tenant_id=? AND id=?", u.TenantID, id)
	return err
}
func (s Service) Touch(ctx context.Context, tenantID, id string) error {
	_, err := s.Store.Exec(ctx, "UPDATE corridors SET active=active,created_at=? WHERE tenant_id=? AND id=?", time.Now().UTC().Format(time.RFC3339Nano), tenantID, id)
	return err
}
