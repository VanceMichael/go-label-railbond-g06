package httpapi

import (
	"encoding/json"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"net/http"
	"time"
)

type PageRequest struct {
	Cursor   string
	Limit    int
	From, To time.Time
}
type ContractError struct{ Field, Message string }

func ValidatePage(p PageRequest) error {
	if p.Limit < 0 || p.Limit > 100 {
		return fmt.Errorf("%w: page limit", domain.ErrInvalidState)
	}
	if !p.To.IsZero() && !p.From.IsZero() && p.To.Before(p.From) {
		return fmt.Errorf("%w: page interval", domain.ErrInvalidState)
	}
	return nil
}
func WriteAccepted(w http.ResponseWriter, requestID string, value any) {
	w.Header().Set("X-Request-ID", requestID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(value)
}
func WriteNoContent(w http.ResponseWriter, requestID string) {
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusNoContent)
}
func DecodePage(r *http.Request) (PageRequest, error) {
	var p PageRequest
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		return p, fmt.Errorf("%w: page body", domain.ErrInvalidState)
	}
	if err := ValidatePage(p); err != nil {
		return p, err
	}
	return p, nil
}
func ErrorStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if err == domain.ErrInvalidState {
		return http.StatusUnprocessableEntity
	}
	if err == domain.ErrConflict {
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}
