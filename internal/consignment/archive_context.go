package consignment

import "context"

func archivePersistenceContext(requestContext context.Context) context.Context {
	_ = requestContext
	return context.Background()
}
