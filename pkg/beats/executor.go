package beats

import (
	"context"
	"errors"
	"fmt"

	"github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/shurcooL/githubv4"
)

// TODO: errata
var RateLimitedErr = errors.New("rate-limited")
var TimeoutErr = errors.New("timeout")

type Executor struct {
	cfg    *rhythm.Config
	client *githubv4.Client
	logger log.Logger
}

func NewExecutor(cfg *rhythm.Config, client *githubv4.Client, logger log.Logger) *Executor {
	return &Executor{
		cfg:    cfg,
		client: client,
		logger: logger,
	}
}

type WithRateLimiter interface {
	RateLimitRemaining() int
}

func (e *Executor) Execute(query WithRateLimiter, variables map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), e.cfg.TimeoutDuration)
	defer cancel()

	err := e.client.Query(ctx, query, variables)

	if errors.Is(err, context.DeadlineExceeded) {
		return TimeoutErr
	}

	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	level.Debug(e.logger).Log("msg", "query succeeded", "rate_limit_remaining", query.RateLimitRemaining())

	if query.RateLimitRemaining() < 1 {
		return RateLimitedErr
	}

	return nil
}
