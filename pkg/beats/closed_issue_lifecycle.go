package beats

import (
	"fmt"

	"github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shurcooL/githubv4"
)

type ClosedIssueLifecycle struct {
	cfg  *rhythm.Config
	exec *Executor

	lifecycleDistribution *prometheus.Desc
}

func (o *ClosedIssueLifecycle) Setup(cfg *rhythm.Config, exec *Executor) {
	o.cfg = cfg
	o.exec = exec

	o.lifecycleDistribution = prometheus.NewDesc("closed_issue_lifecycle", "Distribution of closed issue lifecycles (creation to closed time) by bucket",
		nil, map[string]string{
			"owner": cfg.Owner,
			"repo":  cfg.Repo,
		})
}

func (o *ClosedIssueLifecycle) Collect(ch chan<- prometheus.Metric) {
	// TODO: we might need to collect in the background if there's a lot of pagination happening here,
	// 		 since it might stall the scrape for too long or cause rate-limits to kick in

	type issue struct {
		Id        githubv4.ID
		CreatedAt githubv4.DateTime
		ClosedAt  githubv4.DateTime
	}

	var (
		iterations    uint = 0
		maxIterations uint = 100 // prevent infinite loop in the case of some bug, let's hope there are never more than 100*100 issues
		pageSize      uint = 100
		fetched            = 0

		variables = map[string]interface{}{
			"owner":  githubv4.String(o.cfg.Owner),
			"repo":   githubv4.String(o.cfg.Repo),
			"state":  []githubv4.IssueState{githubv4.IssueStateClosed},
			"cursor": (*githubv4.String)(nil),
			"limit":  githubv4.Int(pageSize),
		}

		sampleCount uint64
		sampleSum   float64

		issues []issue
	)

	buckets := CreateHourBuckets()

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

		iterations++
		err := o.exec.Execute(&query, variables)

		if err != nil {
			// don't export metric upon error; the error is handled by the executor
			return
		}

		issues = append(issues, query.Repository.Issues.Nodes...)

		fetched += len(query.Repository.Issues.Nodes)
		// TODO: logger
		fmt.Println(fetched)

		if !query.Repository.Issues.PageInfo.HasNextPage {
			break
		}

		if iterations > maxIterations {
			panic(fmt.Sprintf("possible infinite loop detected in %T", o))
		}

		variables["cursor"] = githubv4.NewString(query.Repository.Issues.PageInfo.EndCursor)
	}

	for _, issue := range issues {
		hours := issue.ClosedAt.Sub(issue.CreatedAt.Time).Hours()
		sampleCount++
		sampleSum++

		for bucket := range buckets {
			if hours <= bucket {
				buckets[bucket]++
			}
		}
	}

	ch <- prometheus.MustNewConstHistogram(o.lifecycleDistribution, sampleCount, sampleSum, buckets)
}

func (o *ClosedIssueLifecycle) Describe(ch chan<- *prometheus.Desc) {
	ch <- o.lifecycleDistribution
}
