package beats

import (
	"context"
	"errors"

	"github.com/shurcooL/githubv4"
)

// TODO: errata
var RateLimitedErr = errors.New("rate-limited")

type Executor struct {
	client *githubv4.Client
}

func NewExecutor(client *githubv4.Client) *Executor {
	return &Executor{
		client: client,
	}
}

type WithRateLimiter interface {
	RateLimitRemaining() int
}

func (e *Executor) Execute(ctx context.Context, query WithRateLimiter, variables map[string]interface{}) error {
	err := e.client.Query(ctx, query, variables)

	if query.RateLimitRemaining() < 1 {
		return RateLimitedErr
	}

	return err
}
