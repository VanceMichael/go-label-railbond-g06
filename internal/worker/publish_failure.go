package worker

import (
	"context"

	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

func recordPublishFailure(ctx context.Context, store *storage.Store, tenantID, id, owner string, publishErr error) error {
	_, err := store.Exec(ctx, "UPDATE outbox_messages SET last_error=? WHERE tenant_id=? AND id=? AND lease_owner=?", publishErr.Error(), tenantID, id, owner)
	return err
}
