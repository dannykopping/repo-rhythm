package beats

import (
	"time"

	"github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/prometheus/client_golang/prometheus"
)

type Beat interface {
	prometheus.Collector

	Name() string
	Setup(*rhythm.Config, *Executor)
	TickInterval() time.Duration
	Tick() error
}

type Base struct {
	RateLimit RateLimit
}

func (b *Base) RateLimitRemaining() int {
	return b.RateLimit.Remaining
}

type RateLimit struct {
	Remaining int
}

func CreateHourBuckets() map[float64]uint64 {
	return map[float64]uint64{
		// within a day
		time.Hour.Hours():      0,
		6 * time.Hour.Hours():  0,
		24 * time.Hour.Hours(): 0,
		// within a week
		2 * 24 * time.Hour.Hours(): 0,
		4 * 24 * time.Hour.Hours(): 0,
		7 * 24 * time.Hour.Hours(): 0,
		// within a month
		2 * 7 * 24 * time.Hour.Hours(): 0,
		4 * 7 * 24 * time.Hour.Hours(): 0,
		// feckin' old
		60 * 24 * time.Hour.Hours():  0,
		90 * 24 * time.Hour.Hours():  0,
		180 * 24 * time.Hour.Hours(): 0,
	}
}
