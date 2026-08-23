package httpapi_test

import (
	"github.com/VanceMichael/go-base-railbond-g06/internal/httpapi"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBearerAndHeaders(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer token")
	if httpapi.Bearer(r) != "token" {
		t.Fatal("bearer parse")
	}
	r.Header.Set("X-Test", "yes")
	if err := httpapi.RequireHeader(r, "X-Test"); err != nil {
		t.Fatal(err)
	}
}
func TestDecodeRejectsUnknown(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"known":1,"extra":2}`))
	var v struct{ Known int }
	if err := httpapi.DecodeJSON(r, &v); err == nil {
		t.Fatal("unknown field accepted")
	}
}
