package checkpoint

import (
	"context"

	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

func advanceCheckpointProgress(ctx context.Context, tx *storage.Tx, tenantID, consignmentID string, sequence int) error {
	_, err := tx.Exec(ctx, "UPDATE consignments SET status='at_checkpoint',current_checkpoint=?,version=version+1 WHERE tenant_id=? AND id=?", sequence, tenantID, consignmentID)
	return err
}
