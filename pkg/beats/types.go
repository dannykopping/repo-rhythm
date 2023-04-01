package beats

import (
	"context"

	"github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/prometheus/client_golang/prometheus"
)

type Beat interface {
	prometheus.Collector

	Name() string
	Start(context.Context, *rhythm.Config, *Executor) Beat
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
