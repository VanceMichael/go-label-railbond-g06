package consignment

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

type ItemView struct {
	SKU, Description        string
	Quantity, DeclaredValue int
}

func (s Service) Items(ctx context.Context, u domain.User, id string) ([]ItemView, error) {
	rows, err := s.Store.Query(ctx, "SELECT sku,description,quantity,declared_value FROM consignment_items WHERE consignment_id=? ORDER BY id", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ItemView{}
	for rows.Next() {
		var x ItemView
		if err := rows.Scan(&x.SKU, &x.Description, &x.Quantity, &x.DeclaredValue); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, sql.ErrNoRows
	}
	return out, nil
}
func (s Service) MarkInTransit(ctx context.Context, u domain.User, id string) error {
	return s.Advance(ctx, u, id, domain.ConsignmentInTransit)
}
