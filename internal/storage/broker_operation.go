package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
)

// AcquireBrokerOperationKey returns the broker operation key bound to a single
// customs declaration attempt. The key is generated exactly once for a given
// declaration and persisted on the declaration row so that every retry of the
// same remote Release reuses the identical key. This is what lets the external
// broker deduplicate a retried release as a replay of the first attempt rather
// than treating it as a brand new release.
//
// When the declaration already carries a key, that key is returned unchanged.
// Otherwise a new key is minted and written to the row in the same statement,
// guarded by the precondition that the row is in an in-flight state
// (submitted/hold) and has no key yet. A failed precondition (the declaration
// was concurrently released or rejected, or another writer already minted a
// key) is reported as a conflict so the caller does not silently send a second
// distinct release to the broker.
func (s *Store) AcquireBrokerOperationKey(ctx context.Context, tenantID, declarationID, existing string) (string, error) {
	if existing != "" {
		return existing, nil
	}
	key := NewID()
	res, err := s.DB.ExecContext(ctx,
		"UPDATE customs_declarations SET broker_operation_key=? WHERE tenant_id=? AND id=? AND broker_operation_key IS NULL AND status IN ('submitted','hold')",
		key, tenantID, declarationID)
	if err != nil {
		return "", fmt.Errorf("broker operation key: %w", err)
	}
	if n, _ := res.RowsAffected(); n != 1 {
		// Either the declaration no longer exists / is terminal, or another
		// concurrent attempt already minted a key. Either way, refuse to mint
		// a second key: re-read the persisted value so the caller reuses it.
		var stored sql.NullString
		if qerr := s.DB.QueryRowContext(ctx,
			"SELECT broker_operation_key FROM customs_declarations WHERE tenant_id=? AND id=?",
			tenantID, declarationID).Scan(&stored); qerr != nil {
			return "", fmt.Errorf("%w: broker operation key lookup: %v", domain.ErrConflict, qerr)
		}
		if !stored.Valid || stored.String == "" {
			return "", fmt.Errorf("%w: broker operation key not mintable", domain.ErrConflict)
		}
		return stored.String, nil
	}
	return key, nil
}
