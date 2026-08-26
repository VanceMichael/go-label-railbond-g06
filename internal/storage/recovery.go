package storage

import (
	"context"
	"time"
)

const resetExpiredOutboxSQL = "UPDATE outbox_messages SET status='pending',lease_owner=NULL,lease_until=NULL WHERE status='sending' AND lease_until<?"

// ResetExpiredOutbox resets expired outbox leases as a standalone, auto-committed
// statement. Prefer ResetExpiredOutboxTx when the reset must commit atomically
// with other recovery writes (see recovery.Service.RecoverExpired).
func (s *Store) ResetExpiredOutbox(ctx context.Context, now time.Time) (int, error) {
	result, err := s.DB.ExecContext(ctx, resetExpiredOutboxSQL, now.Format(time.RFC3339Nano))
	if err != nil {
		return 0, err
	}
	changed, _ := result.RowsAffected()
	return int(changed), nil
}

// ResetExpiredOutboxTx runs the same reset inside an explicit transaction so it
// commits or rolls back together with the caller's other recovery writes.
func (s *Store) ResetExpiredOutboxTx(ctx context.Context, tx *Tx, now time.Time) (int, error) {
	result, err := tx.Exec(ctx, resetExpiredOutboxSQL, now.Format(time.RFC3339Nano))
	if err != nil {
		return 0, err
	}
	changed, _ := result.RowsAffected()
	return int(changed), nil
}
