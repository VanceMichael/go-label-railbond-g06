package worker

import (
	"math"
	"time"
)

type Backoff struct{ Base, Max time.Duration }

func (b Backoff) Delay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	if b.Base <= 0 {
		b.Base = time.Second
	}
	if b.Max <= 0 {
		b.Max = 30 * time.Minute
	}
	factor := math.Pow(2, float64(attempt))
	d := time.Duration(float64(b.Base) * factor)
	if d > b.Max {
		return b.Max
	}
	return d
}
func NextAttempt(now time.Time, attempt int, b Backoff) time.Time { return now.Add(b.Delay(attempt)) }
func Permanent(attempt, max int) bool                             { return max > 0 && attempt >= max }
