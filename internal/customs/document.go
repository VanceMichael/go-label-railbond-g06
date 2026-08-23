package customs

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"strings"
)

type DocumentCheck struct {
	Kind           string
	Present, Valid bool
	Message        string
}

func CheckDocuments(checks []DocumentCheck) error {
	if len(checks) == 0 {
		return fmt.Errorf("%w: no customs documents", domain.ErrInvalidState)
	}
	for _, c := range checks {
		if strings.TrimSpace(c.Kind) == "" {
			return fmt.Errorf("%w: unnamed customs document", domain.ErrInvalidState)
		}
		if !c.Present || !c.Valid {
			return fmt.Errorf("%w: %s", domain.ErrDeclarationHold, c.Message)
		}
	}
	return nil
}
func RequiredDocuments(status string) []string {
	switch status {
	case "draft", "submitted":
		return []string{"commercial_invoice", "packing_list", "origin_certificate"}
	case "released":
		return []string{"commercial_invoice", "packing_list", "origin_certificate", "release_notice"}
	default:
		return []string{"commercial_invoice"}
	}
}
func (s Service) RequireDocuments(ctx context.Context, u domain.User, id string, checks []DocumentCheck) error {
	if err := CheckDocuments(checks); err != nil {
		return err
	}
	return nil
}
