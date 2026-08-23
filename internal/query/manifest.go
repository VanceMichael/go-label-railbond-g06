package query

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type ManifestQuery struct{ Store *storage.Store }
type ManifestPage struct {
	Items []string
	Next  string
	Total int
}

func (s ManifestQuery) List(ctx context.Context, u domain.User, trainID, status, cursor string, limit int) (ManifestPage, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	args := []any{u.TenantID, trainID}
	where := "c.tenant_id=? AND c.train_id=?"
	if status != "" {
		where += " AND c.status=?"
		args = append(args, status)
	}
	if cursor != "" {
		where += " AND c.id>?"
		args = append(args, cursor)
	}
	var total int
	if err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM consignments c WHERE "+where, args...).Scan(&total); err != nil {
		return ManifestPage{}, err
	}
	args = append(args, limit)
	rows, err := s.Store.Query(ctx, "SELECT c.id FROM consignments c WHERE "+where+" ORDER BY c.id LIMIT ?", args...)
	if err != nil {
		return ManifestPage{}, err
	}
	defer rows.Close()
	out := ManifestPage{Items: []string{}, Total: total}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return ManifestPage{}, err
		}
		out.Items = append(out.Items, id)
	}
	if len(out.Items) == limit {
		out.Next = out.Items[len(out.Items)-1]
	}
	return out, rows.Err()
}
