package beats

import (
	"context"
	"time"

	"github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shurcooL/githubv4"
)

type OpenIssues struct {
	totalCount prometheus.Gauge
}

func (o *OpenIssues) Start(ctx context.Context, cfg *rhythm.Config, exec *Executor) Beat {
	t := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-t.C:
				var query struct {
					Base

					Repository struct {
						Issues struct {
							TotalCount float64
						} `graphql:"issues(states:OPEN)"`
					} `graphql:"repository(name:$repo, owner:$owner)"`
				}

				err := exec.Execute(context.Background(), &query, map[string]interface{}{
					"owner": githubv4.String(cfg.Owner),
					"repo":  githubv4.String(cfg.Repo),
				})
				if err != nil {
					panic(err) // TODO handle errors?
				}

				o.totalCount.Set(query.Repository.Issues.TotalCount)
			case <-ctx.Done():
				return
			}
		}
	}()

	return o
}

func NewOpenIssues() *OpenIssues {
	return &OpenIssues{
		totalCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "open_issues",
			Help: "Current number of open issues",
		}),
	}
}

func (o *OpenIssues) Name() string {
	return "open_issues"
}

func (o *OpenIssues) Collect(ch chan<- prometheus.Metric) {
	o.totalCount.Collect(ch)
}

func (o *OpenIssues) Describe(ch chan<- *prometheus.Desc) {
	o.totalCount.Describe(ch)
}
