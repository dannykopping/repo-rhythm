package beats

import (
	"github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shurcooL/githubv4"
)

type PullRequestCount struct {
	cfg  *rhythm.Config
	exec *Executor

	totalCount *prometheus.Desc
}

func (o *PullRequestCount) Setup(cfg *rhythm.Config, exec *Executor) {
	o.cfg = cfg
	o.exec = exec

	o.totalCount = prometheus.NewDesc("pull_requests", "Current number of pull requests by state",
		[]string{"state"}, map[string]string{
			"owner": cfg.Owner,
			"repo":  cfg.Repo,
		})
}

func (o *PullRequestCount) Collect(ch chan<- prometheus.Metric) {
	var query struct {
		Base

		Repository struct {
			PullRequests struct {
				TotalCount float64
			} `graphql:"pullRequests(states:$state)"`
		} `graphql:"repository(name:$repo, owner:$owner)"`
	}

	states := []githubv4.PullRequestState{githubv4.PullRequestStateOpen, githubv4.PullRequestStateClosed}

	for _, state := range states {
		err := o.exec.Execute(&query, map[string]interface{}{
			"owner": githubv4.String(o.cfg.Owner),
			"repo":  githubv4.String(o.cfg.Repo),
			"state": []githubv4.PullRequestState{state},
		})

		if err != nil {
			// don't export metric upon error; the error is handled by the executor
			continue
		}

		ch <- prometheus.MustNewConstMetric(o.totalCount, prometheus.GaugeValue, query.Repository.PullRequests.TotalCount, string(state))
	}
}

func (o *PullRequestCount) Describe(ch chan<- *prometheus.Desc) {
	ch <- o.totalCount
}
