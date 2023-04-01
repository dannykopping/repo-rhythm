package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/dannykopping/repo-rhythm/pkg/beats"
	repo_rhythm "github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func main() {
	cfg := &repo_rhythm.Config{
		Owner: "grafana",
		Repo:  "loki",
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	exec := beats.NewExecutor(client)

	oi := beats.NewOpenIssues().Start(context.Background(), cfg, exec)

	reg := prometheus.NewRegistry()
	reg.MustRegister(oi)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe("127.0.0.1:9123", nil)) // TODO listen on all addresses
}
