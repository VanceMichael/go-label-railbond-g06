package customs

import (
	"context"
	"fmt"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"strings"
)

type RiskLevel string

const (
	Low    RiskLevel = "low"
	Medium RiskLevel = "medium"
	High   RiskLevel = "high"
)

type RiskSignal struct {
	Code   string
	Weight int
	Reason string
}
type RiskAssessment struct {
	Level   RiskLevel
	Score   int
	Signals []RiskSignal
}

func Assess(items []RiskSignal) RiskAssessment {
	score := 0
	signals := append([]RiskSignal(nil), items...)
	for _, s := range signals {
		if s.Weight > 0 {
			score += s.Weight
		}
	}
	level := Low
	if score >= 70 {
		level = High
	} else if score >= 30 {
		level = Medium
	}
	return RiskAssessment{Level: level, Score: score, Signals: signals}
}
func ValidateAssessment(a RiskAssessment) error {
	if a.Score < 0 {
		return fmt.Errorf("%w: negative customs score", domain.ErrInvalidState)
	}
	if a.Level == High && len(a.Signals) == 0 {
		return fmt.Errorf("%w: high risk without evidence", domain.ErrInvalidState)
	}
	return nil
}
func (s Service) ApplyRisk(ctx context.Context, u domain.User, id string, a RiskAssessment) error {
	if err := ValidateAssessment(a); err != nil {
		return err
	}
	reason := string(a.Level) + ":" + strings.TrimSpace(fmt.Sprint(a.Score))
	if a.Level == High {
		return s.Hold(ctx, u, id, reason)
	}
	return nil
}
