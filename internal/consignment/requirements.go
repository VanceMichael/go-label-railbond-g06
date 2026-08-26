package consignment

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

type Requirement struct {
	Code, Description string
	Satisfied         bool
}

func (s Service) Requirements(ctx context.Context, u domain.User, id string) ([]Requirement, error) {
	var customsStatus string
	if err := s.Store.QueryRow(ctx, "SELECT COALESCE((SELECT status FROM customs_declarations WHERE tenant_id=? AND consignment_id=? LIMIT 1), '')", u.TenantID, id).Scan(&customsStatus); err != nil {
		return nil, err
	}
	var checkpointCount int
	if err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM checkpoint_events WHERE tenant_id=? AND consignment_id=?", u.TenantID, id).Scan(&checkpointCount); err != nil {
		return nil, err
	}
	var docCount int
	if err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM documents WHERE tenant_id=? AND consignment_id=? AND status='sealed'", u.TenantID, id).Scan(&docCount); err != nil {
		return nil, err
	}
	return []Requirement{{"customs_release", "Customs declaration released", customsStatus == "released"}, {"checkpoint_evidence", "Required border evidence", checkpointCount > 0}, {"manifest_document", "Sealed manifest", docCount > 0}}, nil
}
func (s Service) CanDeliver(ctx context.Context, u domain.User, id string) error {
	items, err := s.Requirements(ctx, u, id)
	if err != nil {
		return err
	}
	for _, r := range items {
		if !r.Satisfied {
			return fmt.Errorf("%w: %s", domain.ErrInvalidState, r.Code)
		}
	}
	return nil
}
func (s Service) RecordDeliveryAttempt(ctx context.Context, u domain.User, id string) error {
	_, err := s.Store.Exec(ctx, "UPDATE consignments SET version=version+1 WHERE tenant_id=? AND id=? AND status IN ('in_transit','at_checkpoint')", u.TenantID, id)
	return err
}
