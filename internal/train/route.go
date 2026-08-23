package train

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"strings"
)

type Route struct {
	Origin, Destination string
	Stops               []string
}

func (s Service) Route(ctx context.Context, u domain.User, corridorID string) (Route, error) {
	var r Route
	if err := s.Store.QueryRow(ctx, "SELECT origin,destination FROM corridors WHERE tenant_id=? AND id=?", u.TenantID, corridorID).Scan(&r.Origin, &r.Destination); err != nil {
		return r, err
	}
	rows, err := s.Store.Query(ctx, "SELECT name FROM checkpoints WHERE tenant_id=? AND corridor_id=? ORDER BY sequence_no", u.TenantID, corridorID)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	r.Stops = []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return r, err
		}
		r.Stops = append(r.Stops, name)
	}
	return r, rows.Err()
}
func NormalizeStop(value string) string { return strings.ToUpper(strings.TrimSpace(value)) }
func ValidateRoute(r Route) error {
	if r.Origin == "" || r.Destination == "" || r.Origin == r.Destination {
		return fmt.Errorf("%w: route endpoints", domain.ErrInvalidState)
	}
	return nil
}
func (s Service) SaveRoute(ctx context.Context, u domain.User, id string, r Route) error {
	if err := ValidateRoute(r); err != nil {
		return err
	}
	_, err := s.Store.Exec(ctx, "UPDATE corridors SET origin=?,destination=? WHERE tenant_id=? AND id=?", r.Origin, r.Destination, u.TenantID, id)
	return err
}
