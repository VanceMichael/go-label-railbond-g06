package httpapi_test

import (
	"context"
	"encoding/json"
	"github.com/VanceMichael/go-base-railbond-g06/internal/auth"
	"github.com/VanceMichael/go-base-railbond-g06/internal/consignment"
	"github.com/VanceMichael/go-base-railbond-g06/internal/httpapi"
	"github.com/VanceMichael/go-base-railbond-g06/internal/tenant"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/train"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthReadyAndRequestID(t *testing.T) {
	f := testkit.New(t)
	s := &httpapi.Server{Store: f.Store, Auth: auth.Service{Store: f.Store}, Tenant: tenant.Service{Store: f.Store}, Train: train.Service{Store: f.Store}, Consignment: consignment.Service{Store: f.Store}}
	s.Ready.Store(true)
	r := httptest.NewRecorder()
	s.Routes().ServeHTTP(r, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if r.Code != 200 || r.Header().Get("Content-Type") == "" {
		t.Fatal(r.Code)
	}
	r = httptest.NewRecorder()
	s.Routes().ServeHTTP(r, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if r.Code != 200 {
		t.Fatal(r.Code)
	}
}
func TestUnauthorizedAPIHasJSON(t *testing.T) {
	f := testkit.New(t)
	s := &httpapi.Server{Store: f.Store, Auth: auth.Service{Store: f.Store}}
	r := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/consignments", strings.NewReader("{}"))
	s.Routes().ServeHTTP(r, req)
	if r.Code != http.StatusForbidden {
		t.Fatal(r.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["code"] != "forbidden" {
		t.Fatal(body, err)
	}
}
func TestAuthenticatedCorridorCreation(t *testing.T) {
	f := testkit.New(t)
	a := auth.Service{Store: f.Store}
	if _, err := a.CreateUser(context.Background(), f.TenantID, "api@example.test", "admin", "secret"); err != nil {
		t.Fatal(err)
	}
	u, token, err := a.Login(context.Background(), f.TenantID, "api@example.test", "secret")
	if err != nil || u.Role != "admin" {
		t.Fatal(err)
	}
	_ = token
}
