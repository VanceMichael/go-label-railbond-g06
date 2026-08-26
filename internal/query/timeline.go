package query

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type TimelineRepository struct{ Store *storage.Store }
type TimelineEntry struct{ Kind, Status, Detail string }

func (s TimelineRepository) Load(ctx context.Context, u domain.User, id string) ([]TimelineEntry, error) {
	query, args := checkpointTimelineScope(u.TenantID, id)
	rows, err := s.Store.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []TimelineEntry{}
	for rows.Next() {
		var e TimelineEntry
		if err := rows.Scan(&e.Kind, &e.Status, &e.Detail); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, sql.ErrNoRows
	}
	return out, nil
}
func (s TimelineRepository) Customs(ctx context.Context, u domain.User, id string) (string, error) {
	var st string
	err := s.Store.QueryRow(ctx, "SELECT d.status FROM customs_declarations d JOIN consignments c ON c.id=d.consignment_id WHERE d.tenant_id=? AND c.tenant_id=? AND c.id=?", u.TenantID, u.TenantID, id).Scan(&st)
	return st, err
}
