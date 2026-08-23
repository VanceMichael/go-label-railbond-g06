package checkpoint

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type Event struct {
	ID, CheckpointID, ScannerID, EvidenceHash string
	ObservedAt                                string
}

func (s Service) Events(ctx context.Context, u domain.User, id string) ([]Event, error) {
	rows, err := s.Store.Query(ctx, "SELECT id,checkpoint_id,scanner_id,evidence_hash,observed_at FROM checkpoint_events WHERE tenant_id=? AND consignment_id=? ORDER BY observed_at,id", u.TenantID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Event{}
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.CheckpointID, &e.ScannerID, &e.EvidenceHash, &e.ObservedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
func EnsureTransit(status string) error {
	if status != "in_transit" && status != "at_checkpoint" {
		return domain.ErrInvalidState
	}
	return nil
}

var _ = storage.NewID
