package policy

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type Policy struct {
	MaxDwell          time.Duration
	RequiredDocuments int
	RequireAudit      bool
}
type Evaluator struct{ Store *storage.Store }

func (e Evaluator) ForConsignment(ctx context.Context, u domain.User, id string) (Policy, error) {
	var st string
	if err := e.Store.QueryRow(ctx, "SELECT status FROM consignments WHERE tenant_id=? AND id=?", u.TenantID, id).Scan(&st); err != nil {
		if err == sql.ErrNoRows {
			return Policy{}, domain.ErrNotFound
		}
		return Policy{}, err
	}
	switch st {
	case "booked":
		return Policy{MaxDwell: 48 * time.Hour, RequiredDocuments: 1, RequireAudit: true}, nil
	case "in_transit", "at_checkpoint":
		return Policy{MaxDwell: 24 * time.Hour, RequiredDocuments: 2, RequireAudit: true}, nil
	case "delivered":
		return Policy{MaxDwell: 7 * 24 * time.Hour, RequiredDocuments: 2, RequireAudit: true}, nil
	default:
		return Policy{}, fmt.Errorf("%w: policy for %s", domain.ErrInvalidState, st)
	}
}
func (e Evaluator) MayDepart(ctx context.Context, u domain.User, trainID string) (bool, error) {
	var holds int
	if err := e.Store.QueryRow(ctx, "SELECT COUNT(*) FROM customs_declarations d JOIN consignments c ON c.id=d.consignment_id WHERE c.tenant_id=? AND c.train_id=? AND d.status!='released'", u.TenantID, trainID).Scan(&holds); err != nil {
		return false, err
	}
	return holds == 0, nil
}
func (e Evaluator) MayArchive(ctx context.Context, u domain.User, id string) (bool, error) {
	var invoice, doc int
	if err := e.Store.QueryRow(ctx, "SELECT COUNT(*) FROM invoices WHERE tenant_id=? AND consignment_id=? AND status='settled'", u.TenantID, id).Scan(&invoice); err != nil {
		return false, err
	}
	if err := e.Store.QueryRow(ctx, "SELECT COUNT(*) FROM documents WHERE tenant_id=? AND consignment_id=? AND status='sealed'", u.TenantID, id).Scan(&doc); err != nil {
		return false, err
	}
	return invoice > 0 && doc > 0, nil
}
