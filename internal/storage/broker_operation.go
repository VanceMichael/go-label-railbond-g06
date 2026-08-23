package storage

import "context"

func (s *Store) AcquireBrokerOperationKey(ctx context.Context, tenantID, declarationID, existing string) (string, error) {
	_ = ctx
	_ = tenantID
	_ = declarationID
	if existing != "" {
		return existing, nil
	}
	return NewID(), nil
}
