package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/dannykopping/repo-rhythm/pkg/beats"
	repo_rhythm "github.com/dannykopping/repo-rhythm/pkg/rhythm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func main() {
	// TODO: externally configurable
	cfg := &repo_rhythm.Config{
		Owner:           "grafana",
		Repo:            "loki",
		TimeoutDuration: 10 * time.Second,
		TickInterval:    time.Minute,
	}

	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)

	list := []beats.Beat{
		&beats.Count{},
		&beats.OpenIssueAge{},
		&beats.OpenPullRequestAge{},
		&beats.ClosedIssueLifecycle{},
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	exec := beats.NewExecutor(cfg, client, log.With(logger, "component", "executor"))

	gatherer := prometheus.NewPedanticRegistry()
	reg := prometheus.WrapRegistererWithPrefix("repo_rhythm_", gatherer)

	for _, beat := range list {
		beat.Setup(cfg, exec)
		reg.MustRegister(beat)

		go func(beat beats.Beat) {
			tick := time.NewTicker(cfg.TickInterval)
			defer tick.Stop()

			// tick immediately
			for ; true; <-tick.C {
				start := time.Now()

				log := log.With(logger, "beat", beat.Name(), "interval", cfg.TickInterval)
				level.Info(log).Log("msg", "beat finished", "start", start)

				err := beat.Tick(log)

				if err != nil {
					level.Warn(log).Log("msg", "beat failed", "err", err, start, "duration", time.Since(start))
					continue
				}

				level.Warn(log).Log("msg", "beat succeeded", start, "duration", time.Since(start))
			}
		}(beat)
	}

	http.Handle("/metrics", promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{
		ErrorHandling: promhttp.HTTPErrorOnError,
	}))

	// TODO listen on all addresses
	level.Error(logger).Log("msg", "/metrics handler stopped", "err", http.ListenAndServe("127.0.0.1:9123", nil))
}
