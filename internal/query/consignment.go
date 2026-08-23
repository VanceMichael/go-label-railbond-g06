package query

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

type ConsignmentView struct{ ID, Reference, Status, TrainNumber, ContainerCode string }

func (s OverviewService) Consignment(ctx context.Context, u domain.User, id string) (ConsignmentView, error) {
	var v ConsignmentView
	err := s.Store.QueryRow(ctx, "SELECT c.id,c.reference,c.status,t.number,k.code FROM consignments c JOIN trains t ON t.id=c.train_id JOIN containers k ON k.id=c.container_id WHERE c.tenant_id=? AND c.id=? AND t.tenant_id=? AND k.tenant_id=?", u.TenantID, id, u.TenantID, u.TenantID).Scan(&v.ID, &v.Reference, &v.Status, &v.TrainNumber, &v.ContainerCode)
	if err == sql.ErrNoRows {
		return v, domain.ErrNotFound
	}
	return v, err
}
func (s OverviewService) FindByReference(ctx context.Context, u domain.User, reference string) (ConsignmentView, error) {
	var v ConsignmentView
	err := s.Store.QueryRow(ctx, "SELECT c.id,c.reference,c.status,t.number,k.code FROM consignments c JOIN trains t ON t.id=c.train_id JOIN containers k ON k.id=c.container_id WHERE c.tenant_id=? AND c.reference=?", u.TenantID, reference).Scan(&v.ID, &v.Reference, &v.Status, &v.TrainNumber, &v.ContainerCode)
	return v, err
}
