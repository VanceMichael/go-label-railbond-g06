package consignment

import (
	"context"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

func (s Service) recordCreateAuditAfterCommit(user domain.User, id, requestID, reference string) error {
	ctx := context.Background()
	return s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		return s.Store.RecordAudit(ctx, tx, user.TenantID, user.ID, "consignment.created", "consignment", id, "success", requestID, reference)
	})
}
