package consignment

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type Repository struct{ Store *storage.Store }

func (r Repository) ByReference(ctx context.Context, u domain.User, reference string) (Record, error) {
	var x Record
	err := r.Store.QueryRow(ctx, "SELECT id,reference,status,train_id,container_id FROM consignments WHERE tenant_id=? AND reference=?", u.TenantID, reference).Scan(&x.ID, &x.Reference, &x.Status, &x.TrainID, &x.ContainerID)
	if err == sql.ErrNoRows {
		return x, domain.ErrNotFound
	}
	return x, err
}
func (r Repository) UpdateStatus(ctx context.Context, u domain.User, id, from, to string) (bool, error) {
	res, err := r.Store.Exec(ctx, "UPDATE consignments SET status=?,version=version+1 WHERE tenant_id=? AND id=? AND status=?", to, u.TenantID, id, from)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}
func (r Repository) CountByStatus(ctx context.Context, u domain.User, status string) (int, error) {
	var n int
	err := r.Store.QueryRow(ctx, "SELECT COUNT(*) FROM consignments WHERE tenant_id=? AND status=?", u.TenantID, status).Scan(&n)
	return n, err
}
