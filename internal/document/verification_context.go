package document

import "context"

// manifestVerificationContext returns the request context unchanged so that
// cancellation propagates into the document and consignment-item reads. A
// client that disconnects before the manifest read completes must abort the
// in-flight queries instead of detaching them onto a context that never
// cancels and letting the abandoned call keep holding database resources.
func manifestVerificationContext(requestContext context.Context) context.Context {
	if requestContext == nil {
		return context.Background()
	}
	return requestContext
}
