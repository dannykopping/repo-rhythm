package beats

import (
	"fmt"
	"time"

	"github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shurcooL/githubv4"
)

type OpenPullRequestAge struct {
	cfg  *rhythm.Config
	exec *Executor

	ageDistribution *prometheus.Desc
}

func (o *OpenPullRequestAge) Setup(cfg *rhythm.Config, exec *Executor) {
	o.cfg = cfg
	o.exec = exec

	o.ageDistribution = prometheus.NewDesc("open_pull_request_age", "Distribution of open pull request ages by bucket",
		nil, map[string]string{
			"owner": cfg.Owner,
			"repo":  cfg.Repo,
		})
}

func (o *OpenPullRequestAge) Collect(ch chan<- prometheus.Metric) {
	// TODO: we might need to collect in the background if there's a lot of pagination happening here,
	// 		 since it might stall the scrape for too long or cause rate-limits to kick in

	type pullRequest struct {
		Id        githubv4.ID
		CreatedAt githubv4.DateTime
	}

	var (
		iterations    uint = 0
		maxIterations uint = 100 // prevent infinite loop in the case of some bug, let's hope there are never more than 100*100 pullRequests
		pageSize      uint = 100
		now                = time.Now()
		fetched       int  = 0

		variables = map[string]interface{}{
			"owner":  githubv4.String(o.cfg.Owner),
			"repo":   githubv4.String(o.cfg.Repo),
			"state":  []githubv4.PullRequestState{githubv4.PullRequestStateOpen},
			"cursor": (*githubv4.String)(nil),
			"limit":  githubv4.Int(pageSize),
		}

		sampleCount uint64
		sampleSum   float64

		pullRequests []pullRequest
	)

	buckets := map[float64]uint64{
		// within a day
		time.Hour.Hours():      0,
		6 * time.Hour.Hours():  0,
		24 * time.Hour.Hours(): 0,
		// within a week
		2 * 24 * time.Hour.Hours(): 0,
		4 * 24 * time.Hour.Hours(): 0,
		7 * 24 * time.Hour.Hours(): 0,
		// within a month
		2 * 7 * 24 * time.Hour.Hours(): 0,
		4 * 7 * 24 * time.Hour.Hours(): 0,
		// feckin' old
		60 * 24 * time.Hour.Hours():  0,
		90 * 24 * time.Hour.Hours():  0,
		180 * 24 * time.Hour.Hours(): 0,
	}

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

		iterations++
		err := o.exec.Execute(&query, variables)

		if err != nil {
			// don't export metric upon error; the error is handled by the executor
			return
		}

		pullRequests = append(pullRequests, query.Repository.PullRequests.Nodes...)

		fetched += len(query.Repository.PullRequests.Nodes)
		// TODO: logger
		fmt.Println(fetched)

		if !query.Repository.PullRequests.PageInfo.HasNextPage {
			break
		}

		if iterations > maxIterations {
			panic(fmt.Sprintf("possible infinite loop detected in %T", o))
		}

		variables["cursor"] = githubv4.NewString(query.Repository.PullRequests.PageInfo.EndCursor)
	}

	for _, pullRequest := range pullRequests {
		hours := now.Sub(pullRequest.CreatedAt.Time).Hours()
		sampleCount++
		sampleSum++

		for bucket := range buckets {
			if hours <= bucket {
				buckets[bucket]++
			}
		}
	}

	ch <- prometheus.MustNewConstHistogram(o.ageDistribution, sampleCount, sampleSum, buckets)
}

func (o *OpenPullRequestAge) Describe(ch chan<- *prometheus.Desc) {
	ch <- o.ageDistribution
}
