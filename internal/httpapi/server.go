package httpapi

import (
	"context"
	"encoding/json"
	"github.com/VanceMichael/go-base-railbond-g06/internal/auth"
	"github.com/VanceMichael/go-base-railbond-g06/internal/consignment"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"github.com/VanceMichael/go-base-railbond-g06/internal/tenant"
	"github.com/VanceMichael/go-base-railbond-g06/internal/train"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type Server struct {
	Store       *storage.Store
	Auth        auth.Service
	Tenant      tenant.Service
	Train       train.Service
	Consignment consignment.Service
	Ready       atomic.Bool
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.health)
	mux.HandleFunc("/readyz", s.ready)
	mux.Handle("/api/v1/", s.withRequestID(s.withAuth(http.HandlerFunc(s.api))))
	return mux
}
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
func (s *Server) ready(w http.ResponseWriter, _ *http.Request) {
	if !s.Ready.Load() {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ready"}`))
}
func (s *Server) withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}

type requestIDKey struct{}

func requestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey{}).(string); ok {
		return v
	}
	return ""
}
func (s *Server) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			next.ServeHTTP(w, r)
			return
		}
		raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if raw == "" {
			WriteError(w, domain.ErrForbidden, requestID(r.Context()))
			return
		}
		u, err := s.Auth.Authenticate(r.Context(), raw)
		if err != nil {
			WriteError(w, err, requestID(r.Context()))
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey{}, u)))
	})
}

type userKey struct{}

func currentUser(ctx context.Context) domain.User {
	u, _ := ctx.Value(userKey{}).(domain.User)
	return u
}
func (s *Server) api(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r.Context())
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/")
	switch {
	case r.Method == "POST" && path == "corridors":
		var in struct{ Name, Origin, Destination, Timezone string }
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			WriteError(w, err, requestID(r.Context()))
			return
		}
		id, err := s.Tenant.CreateCorridor(r.Context(), u, in.Name, in.Origin, in.Destination, in.Timezone)
		if err != nil {
			WriteError(w, err, requestID(r.Context()))
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"id": id})
	case r.Method == "POST" && path == "trains":
		var in struct {
			CorridorID, Number string
			Capacity           int
			Departure          time.Time
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			WriteError(w, err, requestID(r.Context()))
			return
		}
		p, err := s.Train.Create(r.Context(), u, in.CorridorID, in.Number, in.Capacity, in.Departure)
		if err != nil {
			WriteError(w, err, requestID(r.Context()))
			return
		}
		writeJSON(w, http.StatusCreated, p)
	case r.Method == "POST" && strings.HasPrefix(path, "trains/") && strings.HasSuffix(path, "/publish"):
		id := strings.TrimSuffix(strings.TrimPrefix(path, "trains/"), "/publish")
		if err := s.Train.Publish(r.Context(), u, id); err != nil {
			WriteError(w, err, requestID(r.Context()))
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "published"})
	case r.Method == "POST" && path == "consignments":
		var in consignment.CreateInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			WriteError(w, err, requestID(r.Context()))
			return
		}
		rec, err := s.Consignment.Create(r.Context(), u, in, requestID(r.Context()))
		if err != nil {
			WriteError(w, err, requestID(r.Context()))
			return
		}
		writeJSON(w, http.StatusCreated, rec)
	case r.Method == "GET" && strings.HasPrefix(path, "consignments/"):
		id := strings.TrimPrefix(path, "consignments/")
		rec, err := s.Consignment.Get(r.Context(), u, id)
		if err != nil {
			WriteError(w, err, requestID(r.Context()))
			return
		}
		writeJSON(w, http.StatusOK, rec)
	default:
		WriteError(w, domain.ErrNotFound, requestID(r.Context()))
	}
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
