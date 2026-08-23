package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

type Service struct {
	Store *storage.Store
	Now   func() time.Time
}

func (s Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now().UTC()
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s Service) CreateTenant(ctx context.Context, name, timezone string) (string, error) {
	id := storage.NewID()
	_, err := s.Store.Exec(ctx, "INSERT INTO tenants(id,name,timezone,created_at) VALUES(?,?,?,?)", id, name, timezone, s.now().Format(time.RFC3339Nano))
	return id, err
}

func (s Service) CreateUser(ctx context.Context, tenantID, email, role, password string) (string, error) {
	if role != "operator" && role != "customs" && role != "finance" && role != "admin" {
		return "", domain.ErrForbidden
	}
	id := storage.NewID()
	_, err := s.Store.Exec(ctx, "INSERT INTO users(id,tenant_id,email,role,password_hash,created_at) VALUES(?,?,?,?,?,?)", id, tenantID, strings.ToLower(email), role, hashToken(password), s.now().Format(time.RFC3339Nano))
	return id, err
}

func (s Service) Login(ctx context.Context, tenantID, email, password string) (domain.User, string, error) {
	var u domain.User
	var stored string
	var active int
	err := s.Store.QueryRow(ctx, "SELECT id,tenant_id,email,role,password_hash,active FROM users WHERE tenant_id=? AND email=?", tenantID, strings.ToLower(email)).Scan(&u.ID, &u.TenantID, &u.Email, &u.Role, &stored, &active)
	if err != nil || active == 0 || stored != hashToken(password) {
		return domain.User{}, "", fmt.Errorf("%w: credentials", domain.ErrForbidden)
	}
	token := storage.NewID()
	_, err = s.Store.Exec(ctx, "INSERT INTO sessions(id,user_id,token_hash,expires_at,created_at) VALUES(?,?,?,?,?)", storage.NewID(), u.ID, hashToken(token), s.now().Add(8*time.Hour).Format(time.RFC3339Nano), s.now().Format(time.RFC3339Nano))
	return u, token, err
}

func (s Service) Authenticate(ctx context.Context, token string) (domain.User, error) {
	var u domain.User
	var exp string
	var revoked interface{}
	err := s.Store.QueryRow(ctx, "SELECT u.id,u.tenant_id,u.email,u.role,s.expires_at,s.revoked_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=?", hashToken(token)).Scan(&u.ID, &u.TenantID, &u.Email, &u.Role, &exp, &revoked)
	if err != nil {
		return u, domain.ErrForbidden
	}
	expiry, e := time.Parse(time.RFC3339Nano, exp)
	if e != nil || !s.now().Before(expiry) || revoked != nil {
		return u, domain.ErrForbidden
	}
	return u, nil
}

func (s Service) Logout(ctx context.Context, token string) error {
	_, err := s.Store.Exec(ctx, "UPDATE sessions SET revoked_at=? WHERE token_hash=? AND revoked_at IS NULL", s.now().Format(time.RFC3339Nano), hashToken(token))
	return err
}
