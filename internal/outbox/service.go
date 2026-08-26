package outbox

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"time"
)

type Service struct{ Store *storage.Store }
type Message struct {
	ID, Topic, AggregateID, Payload string
	Epoch                           int
}

func (s Service) Claim(ctx context.Context, u domain.User, id, owner string, epoch int, until time.Time) error {
	return s.Store.ClaimOutbox(ctx, u.TenantID, id, owner, epoch, until)
}
func (s Service) Acknowledge(ctx context.Context, u domain.User, id, owner string, epoch int) error {
	return s.Store.AckOutbox(ctx, u.TenantID, id, owner, epoch)
}
func (s Service) Fail(ctx context.Context, u domain.User, id, owner string, epoch int, reason string) error {
	return s.Store.FailOutbox(ctx, u.TenantID, id, reason, time.Now().UTC().Add(time.Minute))
}
func (s Service) List(ctx context.Context, u domain.User) ([]Message, error) {
	rows, err := s.Store.Query(ctx, "SELECT id,topic,aggregate_id,payload,lease_epoch FROM outbox_messages WHERE tenant_id=? AND status IN ('pending','sending') ORDER BY created_at", u.TenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Message{}
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.Topic, &m.AggregateID, &m.Payload, &m.Epoch); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func IsMissing(err error) bool { return err == sql.ErrNoRows }
func Wrap(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("outbox: %w", err)
}
