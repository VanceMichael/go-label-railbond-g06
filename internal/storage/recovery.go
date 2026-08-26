package storage

import (
	"context"
	"time"
)

func (s *Store) ResetExpiredOutbox(ctx context.Context, now time.Time) (int, error) {
	result, err := s.DB.ExecContext(ctx, "UPDATE outbox_messages SET status='pending',lease_owner=NULL,lease_until=NULL WHERE status='sending' AND lease_until<?", now.Format(time.RFC3339Nano))
	if err != nil {
		return 0, err
	}
	changed, _ := result.RowsAffected()
	return int(changed), nil
}
