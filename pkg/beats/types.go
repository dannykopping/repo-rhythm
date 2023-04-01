package beats

import (
	"github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/prometheus/client_golang/prometheus"
)

type Beat interface {
	prometheus.Collector

	Setup(*rhythm.Config, *Executor)
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
