package storage

import (
	"context"
	"database/sql"
)

// ObserveRebookReplay looks up a previously completed rebook for the given
// idempotency key. When a prior completion exists the stored response body
// (the original route assignment id) is returned together with replay=true so
// the caller can converge onto the first result instead of creating a second
// assignment.
func (s *Store) ObserveRebookReplay(ctx context.Context, tx *Tx, tenantID, key string) (string, bool, error) {
	var storedResponse string
	err := tx.QueryRow(ctx, "SELECT response_body FROM idempotency_keys WHERE tenant_id=? AND key=? AND method='POST' AND path='/rebook'", tenantID, key).Scan(&storedResponse)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return storedResponse, true, nil
}
