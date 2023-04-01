package beats

import (
	"github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shurcooL/githubv4"
)

type IssueCount struct {
	cfg  *rhythm.Config
	exec *Executor

	totalCount *prometheus.Desc
}

func (o *IssueCount) Setup(cfg *rhythm.Config, exec *Executor) {
	o.cfg = cfg
	o.exec = exec

	o.totalCount = prometheus.NewDesc("issues", "Current number of issues by state",
		[]string{"state"}, map[string]string{
			"owner": cfg.Owner,
			"repo":  cfg.Repo,
		})
}

func (o *IssueCount) Collect(ch chan<- prometheus.Metric) {
	var query struct {
		Base

		Repository struct {
			Issues struct {
				TotalCount float64
			} `graphql:"issues(states:$state)"`
		} `graphql:"repository(name:$repo, owner:$owner)"`
	}

	states := []githubv4.IssueState{githubv4.IssueStateOpen, githubv4.IssueStateClosed}

	for _, state := range states {
		err := o.exec.Execute(&query, map[string]interface{}{
			"owner": githubv4.String(o.cfg.Owner),
			"repo":  githubv4.String(o.cfg.Repo),
			"state": []githubv4.IssueState{state},
		})

		if err != nil {
			// don't export metric upon error; the error is handled by the executor
			continue
		}

		ch <- prometheus.MustNewConstMetric(o.totalCount, prometheus.GaugeValue, query.Repository.Issues.TotalCount, string(state))
	}
}

func (o *IssueCount) Describe(ch chan<- *prometheus.Desc) {
	ch <- o.totalCount
}
