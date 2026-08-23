package query

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"strings"
)

type SearchResult struct{ ID, Reference, Status string }

func (s ManifestQuery) Search(ctx context.Context, u domain.User, term string, limit int) ([]SearchResult, error) {
	term = strings.TrimSpace(term)
	if term == "" {
		return nil, sql.ErrNoRows
	}
	if limit < 1 {
		limit = 20
	}
	rows, err := s.Store.Query(ctx, "SELECT id,reference,status FROM consignments WHERE tenant_id=? AND (reference LIKE ? OR id LIKE ?) ORDER BY reference,id LIMIT ?", u.TenantID, "%"+term+"%", "%"+term+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SearchResult{}
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.ID, &r.Reference, &r.Status); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
func (s ManifestQuery) Statuses(ctx context.Context, u domain.User) (map[string]int, error) {
	rows, err := s.Store.Query(ctx, "SELECT status,COUNT(*) FROM consignments WHERE tenant_id=? GROUP BY status", u.TenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var st string
		var n int
		if err := rows.Scan(&st, &n); err != nil {
			return nil, err
		}
		out[st] = n
	}
	return out, rows.Err()
}
