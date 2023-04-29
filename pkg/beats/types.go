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

func CreateDayBuckets() map[string]float64 {
	day := 24 * time.Hour.Hours()
	week := 7 * day
	month := 30 * day
	year := 365 * day

	return map[string]float64{
		"1d":   day,
		"2d":   2 * day,
		"4d":   4 * day,
		"7d":   week,
		"14d":  2 * week,
		"30d":  month,
		"60d":  2 * month,
		"90d":  3 * month,
		"180d": 6 * month,
		"365d": year,
		"730d": 2 * year,
	}
}
