package customs

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"time"
)

type BrokerStatus struct {
	OperationKey, Status string
	CheckedAt            time.Time
}
type BrokerGateway interface {
	Status(context.Context, string) (BrokerStatus, error)
}

func ReconcileBroker(ctx context.Context, g BrokerGateway, key string) (BrokerStatus, error) {
	if key == "" {
		return BrokerStatus{}, fmt.Errorf("%w: broker operation key", domain.ErrInvalidState)
	}
	if err := domain.CheckContext(ctx); err != nil {
		return BrokerStatus{}, err
	}
	if g == nil {
		return BrokerStatus{}, fmt.Errorf("broker gateway unavailable")
	}
	s, err := g.Status(ctx, key)
	if err != nil {
		return BrokerStatus{}, fmt.Errorf("broker status: %w", err)
	}
	if s.Status != "released" && s.Status != "submitted" && s.Status != "hold" {
		return s, fmt.Errorf("%w: broker status %s", domain.ErrInvalidState, s.Status)
	}
	return s, nil
}

type StaticGateway struct{ Value BrokerStatus }

func (g StaticGateway) Status(context.Context, string) (BrokerStatus, error) { return g.Value, nil }
