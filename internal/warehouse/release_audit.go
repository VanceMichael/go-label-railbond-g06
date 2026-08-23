package warehouse

import (
	"context"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

func (s Service) recordCustomsReleaseAuditAfterCommit(user domain.User, consignmentID, requestID string) error {
	ctx := context.Background()
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		return s.Store.RecordAudit(ctx, tx, user.TenantID, user.ID, "warehouse.released", "consignment", consignmentID, "success", requestID, "customs rejection")
	})
}
