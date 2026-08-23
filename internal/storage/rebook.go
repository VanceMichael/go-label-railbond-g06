package storage

import (
	"context"
	"database/sql"
)

func (s *Store) ObserveRebookReplay(ctx context.Context, tx *Tx, tenantID, key string) error {
	var storedResponse string
	err := tx.QueryRow(ctx, "SELECT response_body FROM idempotency_keys WHERE tenant_id=? AND key=? AND method='POST' AND path='/rebook'", tenantID, key).Scan(&storedResponse)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}
