package httpapi

import (
	"encoding/json"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"net/http"
)

type ErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func WriteError(w http.ResponseWriter, err error, requestID string) {
	status := http.StatusInternalServerError
	code := "internal_error"
	msg := "internal server error"
	switch {
	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
		code = "forbidden"
		msg = "access denied"
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		code = "not_found"
		msg = "resource not found"
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrLeaseLost):
		status = http.StatusConflict
		code = "conflict"
		msg = "resource conflict"
	case errors.Is(err, domain.ErrDeclarationHold):
		status = http.StatusConflict
		code = "declaration_hold"
		msg = "customs declaration is on hold"
	case errors.Is(err, domain.ErrInvalidState):
		status = http.StatusUnprocessableEntity
		code = "invalid_state"
		msg = "business state does not allow this operation"
	case errors.Is(err, domain.ErrCancelled):
		status = http.StatusRequestTimeout
		code = "cancelled"
		msg = "request was cancelled"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorBody{Code: code, Message: msg, RequestID: requestID})
}
