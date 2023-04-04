package beats

import (
	"time"

	"github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shurcooL/githubv4"
)

type Count struct {
	cfg  *rhythm.Config
	exec *Executor

	issueCount       *prometheus.GaugeVec
	pullRequestCount *prometheus.GaugeVec
}

func (o *Count) Name() string {
	return "count issues & PRs"
}

func (o *Count) TickInterval() time.Duration {
	return time.Minute
}

func (o *Count) Setup(cfg *rhythm.Config, exec *Executor) {
	o.cfg = cfg
	o.exec = exec

	o.issueCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "issues",
		Help: "Current number of issues by state",
		ConstLabels: map[string]string{
			"owner": cfg.Owner,
			"repo":  cfg.Repo,
		},
	}, []string{"state"})
	o.pullRequestCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pull_requests",
		Help: "Current number of pull requests by state",
		ConstLabels: map[string]string{
			"owner": cfg.Owner,
			"repo":  cfg.Repo,
		},
	}, []string{"state"})
}

func (o *Count) Tick() error {
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
			return err
		}

		o.issueCount.WithLabelValues(string(issueState)).Set(query.Repository.Issues.TotalCount)
		o.pullRequestCount.WithLabelValues(string(prState)).Set(query.Repository.PullRequests.TotalCount)
	}

	return nil
}

func (o *Count) Collect(ch chan<- prometheus.Metric) {
	o.issueCount.Collect(ch)
	o.pullRequestCount.Collect(ch)
}

func (o *Count) Describe(ch chan<- *prometheus.Desc) {
	o.issueCount.Describe(ch)
	o.pullRequestCount.Describe(ch)
}
