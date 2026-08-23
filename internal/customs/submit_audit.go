package customs

import (
	"context"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

func (s Service) recordSubmitAuditAfterCommit(user domain.User, id, requestID, consignment string) error {
	ctx := context.Background()
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		return s.Store.RecordAudit(ctx, tx, user.TenantID, user.ID, "customs.submitted", "declaration", id, "success", requestID, consignment)
	})
}
