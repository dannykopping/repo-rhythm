package beats

import (
	"github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shurcooL/githubv4"
)

type Count struct {
	cfg  *rhythm.Config
	exec *Executor

	issueCount       *prometheus.Desc
	pullRequestCount *prometheus.Desc
}

func (o *Count) Setup(cfg *rhythm.Config, exec *Executor) {
	o.cfg = cfg
	o.exec = exec

	o.issueCount = prometheus.NewDesc("issues", "Current number of issues by state",
		[]string{"state"}, map[string]string{
			"owner": cfg.Owner,
			"repo":  cfg.Repo,
		})
	o.pullRequestCount = prometheus.NewDesc("pull_requests", "Current number of pull requests by state",
		[]string{"state"}, map[string]string{
			"owner": cfg.Owner,
			"repo":  cfg.Repo,
		})
}

func (o *Count) Collect(ch chan<- prometheus.Metric) {
	var query struct {
		Base

		Repository struct {
			Issues struct {
				TotalCount float64
			} `graphql:"issues(states:$issueState)"`
			PullRequests struct {
				TotalCount float64
			} `graphql:"pullRequests(states:$prState)"`
		} `graphql:"repository(name:$repo, owner:$owner)"`
	}

	issueStates := []githubv4.IssueState{githubv4.IssueStateOpen, githubv4.IssueStateClosed}
	prStates := []githubv4.PullRequestState{githubv4.PullRequestStateOpen, githubv4.PullRequestStateClosed}

	for i := 0; i < len(issueStates); i++ {
		issueState := issueStates[i]
		prState := prStates[i]
		err := o.exec.Execute(&query, map[string]interface{}{
			"owner":      githubv4.String(o.cfg.Owner),
			"repo":       githubv4.String(o.cfg.Repo),
			"issueState": []githubv4.IssueState{issueState},
			"prState":    []githubv4.PullRequestState{prState},
		})

		if err != nil {
			// don't export metric upon error; the error is handled by the executor
			continue
		}

		ch <- prometheus.MustNewConstMetric(o.issueCount, prometheus.GaugeValue, query.Repository.Issues.TotalCount, string(issueState))
		ch <- prometheus.MustNewConstMetric(o.pullRequestCount, prometheus.GaugeValue, query.Repository.PullRequests.TotalCount, string(prState))
	}
}

func (o *Count) Describe(ch chan<- *prometheus.Desc) {
	ch <- o.issueCount
	ch <- o.pullRequestCount
}
