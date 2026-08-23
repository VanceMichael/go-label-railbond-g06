package httpapi

import (
	"encoding/json"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"io"
	"net/http"
	"strings"
)

func DecodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return fmt.Errorf("%w: empty body", domain.ErrInvalidState)
	}
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%w: request body: %v", domain.ErrInvalidState, err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("%w: multiple JSON values", domain.ErrInvalidState)
	}
	return nil
}
func Bearer(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(value) < 7 || !strings.EqualFold(value[:6], "bearer") {
		return ""
	}
	return strings.TrimSpace(value[6:])
}
func RequireHeader(r *http.Request, name string) error {
	if strings.TrimSpace(r.Header.Get(name)) == "" {
		return fmt.Errorf("%w: header %s", domain.ErrInvalidState, name)
	}
	return nil
}
