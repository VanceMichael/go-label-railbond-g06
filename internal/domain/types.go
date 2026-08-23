package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound        = errors.New("railbond: not found")
	ErrConflict        = errors.New("railbond: conflict")
	ErrForbidden       = errors.New("railbond: forbidden")
	ErrInvalidState    = errors.New("railbond: invalid state")
	ErrDeclarationHold = errors.New("railbond: declaration hold")
	ErrLeaseLost       = errors.New("railbond: lease lost")
	ErrCancelled       = errors.New("railbond: operation cancelled")
)

func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}

func CheckContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w: %v", ErrCancelled, ctx.Err())
	default:
		return nil
	}
}

type ConsignmentStatus string

const (
	ConsignmentDraft        ConsignmentStatus = "draft"
	ConsignmentBooked       ConsignmentStatus = "booked"
	ConsignmentInTransit    ConsignmentStatus = "in_transit"
	ConsignmentAtCheckpoint ConsignmentStatus = "at_checkpoint"
	ConsignmentDelivered    ConsignmentStatus = "delivered"
	ConsignmentArchived     ConsignmentStatus = "archived"
)

func (s ConsignmentStatus) CanMove(to ConsignmentStatus) bool {
	switch s {
	case ConsignmentDraft:
		return to == ConsignmentBooked
	case ConsignmentBooked:
		return to == ConsignmentInTransit
	case ConsignmentInTransit:
		return to == ConsignmentAtCheckpoint || to == ConsignmentDelivered
	case ConsignmentAtCheckpoint:
		return to == ConsignmentInTransit || to == ConsignmentDelivered
	case ConsignmentDelivered:
		return to == ConsignmentArchived
	default:
		return false
	}
}

type TrainStatus string

const (
	TrainPlanned   TrainStatus = "planned"
	TrainPublished TrainStatus = "published"
	TrainDeparted  TrainStatus = "departed"
	TrainArrived   TrainStatus = "arrived"
	TrainCancelled TrainStatus = "cancelled"
)

func (s TrainStatus) CanMove(to TrainStatus) bool {
	switch s {
	case TrainPlanned:
		return to == TrainPublished || to == TrainCancelled
	case TrainPublished:
		return to == TrainDeparted || to == TrainCancelled
	case TrainDeparted:
		return to == TrainArrived
	default:
		return false
	}
}

type DeclarationStatus string

const (
	DeclarationDraft     DeclarationStatus = "draft"
	DeclarationSubmitted DeclarationStatus = "submitted"
	DeclarationHold      DeclarationStatus = "hold"
	DeclarationReleased  DeclarationStatus = "released"
	DeclarationRejected  DeclarationStatus = "rejected"
)

func (s DeclarationStatus) CanMove(to DeclarationStatus) bool {
	switch s {
	case DeclarationDraft:
		return to == DeclarationSubmitted || to == DeclarationRejected
	case DeclarationSubmitted:
		return to == DeclarationReleased || to == DeclarationHold || to == DeclarationRejected
	case DeclarationHold:
		return to == DeclarationSubmitted || to == DeclarationRejected
	default:
		return false
	}
}

type InvoiceStatus string

const (
	InvoiceIssued   InvoiceStatus = "issued"
	InvoiceDisputed InvoiceStatus = "disputed"
	InvoiceSettled  InvoiceStatus = "settled"
)

type Lease struct {
	Owner string
	Token string
	Epoch int
	Until time.Time
}

func (l Lease) Active(now time.Time) bool { return l.Owner != "" && now.Before(l.Until) }

type User struct {
	ID       string
	TenantID string
	Email    string
	Role     string
}

func RequireRole(user User, roles ...string) error {
	for _, role := range roles {
		if user.Role == role {
			return nil
		}
	}
	return ErrForbidden
}
