package customs

import "fmt"

func unresolvedBrokerFailure(declarationID string, releaseErr error) error {
	return fmt.Errorf("broker release %s: %w", declarationID, releaseErr)
}
