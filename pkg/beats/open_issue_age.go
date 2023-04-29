package beats

import (
	"fmt"
	"time"

	"github.com/dannykopping/repo-rhythm/pkg/metrics"
	"github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shurcooL/githubv4"
)

type OpenIssueAge struct {
	cfg  *rhythm.Config
	exec *Executor

	age metrics.Distribution
}

func (o *OpenIssueAge) Name() string {
	return "open issues age"
}

func (o *OpenIssueAge) TickInterval() time.Duration {
	return time.Hour
}

func (o *OpenIssueAge) Setup(cfg *rhythm.Config, exec *Executor) {
	o.cfg = cfg
	o.exec = exec
	o.age = metrics.NewDistribution(
		metrics.DistributionOpts{
			Name: "open_issue_age",
			Help: "Distribution of open issue ages by days",
			ConstLabels: map[string]string{
				"owner": cfg.Owner,
				"repo":  cfg.Repo,
			},
		},
		CreateDayBuckets(),
	)
}

func (o *OpenIssueAge) Tick() error {
	type issue struct {
		Id        githubv4.ID
		CreatedAt githubv4.DateTime
	}

	var (
		pageSize uint = 100
		now           = time.Now()
		fetched       = 0

		variables = map[string]interface{}{
			"owner":  githubv4.String(o.cfg.Owner),
			"repo":   githubv4.String(o.cfg.Repo),
			"state":  []githubv4.IssueState{githubv4.IssueStateOpen},
			"cursor": (*githubv4.String)(nil),
			"limit":  githubv4.Int(pageSize),
		}

		issues []issue
	)

	for {
		var query struct {
			Base

			Repository struct {
				Issues struct {
					Nodes []issue

					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"issues(states:$state, first:$limit, after:$cursor)"`
			} `graphql:"repository(name:$repo, owner:$owner)"`
		}

		err := o.exec.Execute(&query, variables)
		if err != nil {
			// don't export metric upon error; the error is handled by the executor
			return err
		}

		issues = append(issues, query.Repository.Issues.Nodes...)
		fetched += len(query.Repository.Issues.Nodes)

		// TODO: logger
		fmt.Println(fetched)

		if !query.Repository.Issues.PageInfo.HasNextPage {
			break
		}

		variables["cursor"] = githubv4.NewString(query.Repository.Issues.PageInfo.EndCursor)
	}

	o.age.Reset()
	for _, issue := range issues {
		hours := now.Sub(issue.CreatedAt.Time)
		o.age.Observe(hours.Hours())
	}

	return nil
}

func (o *OpenIssueAge) Collect(ch chan<- prometheus.Metric) {
	o.age.Collect(ch)
}

func (o *OpenIssueAge) Describe(ch chan<- *prometheus.Desc) {
	o.age.Describe(ch)
}
