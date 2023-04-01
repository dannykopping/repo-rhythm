package beats

import (
	"github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shurcooL/githubv4"
)

type OpenIssues struct {
	cfg  *rhythm.Config
	exec *Executor

	totalCount *prometheus.Desc
}

func (o *OpenIssues) Setup(cfg *rhythm.Config, exec *Executor) {
	o.cfg = cfg
	o.exec = exec

	o.totalCount = prometheus.NewDesc("open_issues", "Current number of open issues", nil, map[string]string{
		"owner": cfg.Owner,
		"repo":  cfg.Repo,
	})
}

func (o *OpenIssues) Collect(ch chan<- prometheus.Metric) {
	var query struct {
		Base

		Repository struct {
			Issues struct {
				TotalCount float64
			} `graphql:"issues(states:OPEN)"`
		} `graphql:"repository(name:$repo, owner:$owner)"`
	}

	err := o.exec.Execute(&query, map[string]interface{}{
		"owner": githubv4.String(o.cfg.Owner),
		"repo":  githubv4.String(o.cfg.Repo),
	})

	if err != nil {
		// don't export metric upon error; the error is handled by the executor
		return
	}

	ch <- prometheus.MustNewConstMetric(o.totalCount, prometheus.GaugeValue, query.Repository.Issues.TotalCount)
}

func (o *OpenIssues) Describe(ch chan<- *prometheus.Desc) {
	ch <- o.totalCount
}
