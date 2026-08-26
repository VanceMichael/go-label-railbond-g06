package customs

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

type QueueItem struct{ ID, ConsignmentID, Status, Reason string }

func (s Service) Queue(ctx context.Context, u domain.User, status string) ([]QueueItem, error) {
	rows, err := s.Store.Query(ctx, "SELECT id,consignment_id,status,COALESCE(hold_reason,'') FROM customs_declarations WHERE tenant_id=? AND status=? ORDER BY id", u.TenantID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []QueueItem{}
	for rows.Next() {
		var x QueueItem
		if err := rows.Scan(&x.ID, &x.ConsignmentID, &x.Status, &x.Reason); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s Service) HasRelease(ctx context.Context, u domain.User, consignmentID string) (bool, error) {
	var n int
	err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM customs_declarations WHERE tenant_id=? AND consignment_id=? AND status='released'", u.TenantID, consignmentID).Scan(&n)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return n > 0, err
}
