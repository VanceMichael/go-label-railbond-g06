package manifest

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"sort"
)

type Service struct{ Store *storage.Store }
type Entry struct {
	ConsignmentID, Reference, Status string
	ItemCount                        int
}

func (s Service) Build(ctx context.Context, u domain.User, trainID string) ([]Entry, error) {
	rows, err := s.Store.Query(ctx, "SELECT c.id,c.reference,c.status,COUNT(i.id) FROM consignments c LEFT JOIN consignment_items i ON i.consignment_id=c.id WHERE c.tenant_id=? AND c.train_id=? GROUP BY c.id,c.reference,c.status ORDER BY c.reference,c.id", u.TenantID, trainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Entry{}
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ConsignmentID, &e.Reference, &e.Status, &e.ItemCount); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
func (s Service) Validate(ctx context.Context, u domain.User, trainID string) error {
	items, err := s.Build(ctx, u, trainID)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return fmt.Errorf("%w: empty manifest", domain.ErrInvalidState)
	}
	for _, i := range items {
		if i.Status == "archived" {
			return fmt.Errorf("%w: archived manifest item", domain.ErrInvalidState)
		}
		if i.ItemCount == 0 {
			return fmt.Errorf("%w: itemless consignment", domain.ErrInvalidState)
		}
	}
	return nil
}
func SortEntries(items []Entry) []Entry {
	out := append([]Entry(nil), items...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Reference == out[j].Reference {
			return out[i].ConsignmentID < out[j].ConsignmentID
		}
		return out[i].Reference < out[j].Reference
	})
	return out
}
