package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dannykopping/repo-rhythm/pkg/beats"
	repo_rhythm "github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func main() {
	// TODO: externally configurable
	cfg := &repo_rhythm.Config{
		Owner:           "grafana",
		Repo:            "loki",
		TimeoutDuration: 10 * time.Second,
	}

	list := []beats.Beat{
		&beats.IssueCount{},
		&beats.OpenIssueAge{},
		&beats.ClosedIssueLifecycle{},
		&beats.PullRequestCount{},
		&beats.OpenPullRequestAge{},
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	exec := beats.NewExecutor(cfg, client)

	gatherer := prometheus.NewPedanticRegistry()
	reg := prometheus.WrapRegistererWithPrefix("repo_rhythm_", gatherer)

	for _, beat := range list {
		beat.Setup(cfg, exec)
		reg.MustRegister(beat)
	}

	http.Handle("/metrics", promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe("127.0.0.1:9123", nil)) // TODO listen on all addresses
}
