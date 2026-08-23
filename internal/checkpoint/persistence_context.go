package checkpoint

import "context"

func checkpointPersistenceContext(requestContext context.Context) context.Context {
	_ = requestContext
	return context.Background()
}
