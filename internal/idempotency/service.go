package idempotency

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
)

var ErrKeyMismatch = errors.New("idempotency: request key belongs to another operation")

type Result struct {
	Status int
	Body   string
	Replay bool
}
type Service struct{ Store *storage.Store }

func (s Service) Execute(ctx context.Context, u domain.User, key, method, path string, operation func(*storage.Tx) (int, string, error)) (Result, error) {
	if key == "" || method == "" || path == "" {
		return Result{}, fmt.Errorf("%w: idempotency input", domain.ErrInvalidState)
	}
	var out Result
	err := s.Store.WithTx(ctx, func(tx *storage.Tx) error {
		var status int
		var body, storedMethod, storedPath string
		err := tx.QueryRow(ctx, "SELECT status_code,response_body,method,path FROM idempotency_keys WHERE tenant_id=? AND key=?", u.TenantID, key).Scan(&status, &body, &storedMethod, &storedPath)
		if err == nil {
			if storedMethod != method || storedPath != path {
				return ErrKeyMismatch
			}
			out = Result{Status: status, Body: body, Replay: true}
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		code, response, err := operation(tx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "INSERT INTO idempotency_keys(id,tenant_id,key,method,path,status_code,response_body,created_at) VALUES(?,?,?,?,?,?,?,datetime('now'))", storage.NewID(), u.TenantID, key, method, path, code, response); err != nil {
			return err
		}
		out = Result{Status: code, Body: response}
		return nil
	})
	return out, err
}
func (s Service) Forget(ctx context.Context, u domain.User, key string) error {
	_, err := s.Store.Exec(ctx, "DELETE FROM idempotency_keys WHERE tenant_id=? AND key=?", u.TenantID, key)
	return err
}
func (s Service) Count(ctx context.Context, u domain.User) (int, error) {
	var n int
	err := s.Store.QueryRow(ctx, "SELECT COUNT(*) FROM idempotency_keys WHERE tenant_id=?", u.TenantID).Scan(&n)
	return n, err
}
