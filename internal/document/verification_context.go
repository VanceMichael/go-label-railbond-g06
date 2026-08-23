package document

import "context"

func manifestVerificationContext(requestContext context.Context) context.Context {
	_ = requestContext
	return context.Background()
}
