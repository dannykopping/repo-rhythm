package beats

import (
	"fmt"
	"time"

	"github.com/dannykopping/repo-rhythm/pkg/metrics"
	"github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shurcooL/githubv4"
)

type OpenPullRequestAge struct {
	cfg  *rhythm.Config
	exec *Executor

	age metrics.Distribution
}

func (o *OpenPullRequestAge) Name() string {
	return "open pull requests age"
}

func (o *OpenPullRequestAge) Setup(cfg *rhythm.Config, exec *Executor) {
	o.cfg = cfg
	o.exec = exec

	o.age = metrics.NewDistribution(
		metrics.DistributionOpts{
			Name: "open_pull_request_age",
			Help: "Distribution of open pull request ages by days",
			ConstLabels: map[string]string{
				"owner": cfg.Owner,
				"repo":  cfg.Repo,
			},
		},
		CreateDayBuckets(),
	)
}

func (o *OpenPullRequestAge) Tick(log.Logger) error {
	type pullRequest struct {
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
			"state":  []githubv4.PullRequestState{githubv4.PullRequestStateOpen},
			"cursor": (*githubv4.String)(nil),
			"limit":  githubv4.Int(pageSize),
		}

		pullRequests []pullRequest
	)

	for {
		var query struct {
			Base

			Repository struct {
				PullRequests struct {
					Nodes []pullRequest

					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"pullRequests(states:$state, first:$limit, after:$cursor)"`
			} `graphql:"repository(name:$repo, owner:$owner)"`
		}

		err := o.exec.Execute(&query, variables)
		if err != nil {
			// don't export metric upon error; the error is handled by the executor
			return err
		}

		pullRequests = append(pullRequests, query.Repository.PullRequests.Nodes...)

		fetched += len(query.Repository.PullRequests.Nodes)
		// TODO: logger
		fmt.Println(fetched)

		if !query.Repository.PullRequests.PageInfo.HasNextPage {
			break
		}

		variables["cursor"] = githubv4.NewString(query.Repository.PullRequests.PageInfo.EndCursor)
	}

	o.age.Reset()
	for _, issue := range pullRequests {
		hours := now.Sub(issue.CreatedAt.Time)
		o.age.Observe(hours.Hours())
	}

	return nil
}

func (o *OpenPullRequestAge) Collect(ch chan<- prometheus.Metric) {
	o.age.Collect(ch)
}

func (o *OpenPullRequestAge) Describe(ch chan<- *prometheus.Desc) {
	o.age.Describe(ch)
}
