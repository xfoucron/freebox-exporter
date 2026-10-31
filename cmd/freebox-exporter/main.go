package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/xfoucron/freebox-exporter/internal/collector"
	"github.com/xfoucron/freebox-exporter/internal/freebox"
)

const (
	appID   = "freebox-exporter"
	appName = "Freebox Exporter"

	maxConcurrentScrapes = 2
)

var version = "dev"

type options struct {
	freeboxURL  string
	listen      string
	dataDir     string
	timeout     time.Duration
	showVersion bool
}

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := run(os.Args[1:], log); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(args []string, log *slog.Logger) error {
	var opts options
	fs := flag.NewFlagSet("freebox-exporter", flag.ContinueOnError)
	fs.StringVar(&opts.freeboxURL, "freebox.url", envOr("FREEBOX_URL", "http://mafreebox.freebox.fr"), "base URL of the Freebox")
	fs.StringVar(&opts.listen, "listen", envOr("FREEBOX_LISTEN", ":9091"), "address to listen on")
	fs.StringVar(&opts.dataDir, "data-dir", envOr("FREEBOX_DATA_DIR", defaultDataDir()), "directory where the application token is stored")
	fs.DurationVar(&opts.timeout, "scrape-timeout", 10*time.Second, "timeout of a single scrape")
	fs.BoolVar(&opts.showVersion, "version", false, "print the version and exit")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if opts.showVersion {
		fmt.Println(version)
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return serve(ctx, opts, log)
}

func serve(ctx context.Context, opts options, log *slog.Logger) error {
	token, err := resolveToken(ctx, opts, log)
	if err != nil {
		return err
	}
	client, err := freebox.New(ctx, opts.freeboxURL, appID, token, nil)
	if err != nil {
		return err
	}

	registry := prometheus.NewRegistry()
	buildInfo := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "freebox_exporter_build_info",
		Help:        "Build information of the exporter.",
		ConstLabels: prometheus.Labels{"version": version},
	})
	buildInfo.Set(1)
	registry.MustRegister(
		buildInfo,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collector.New(client, opts.timeout, log),
	)

	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry, MaxRequestsInFlight: maxConcurrentScrapes}))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	srv := &http.Server{
		Addr:              opts.listen,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Info("listening", "addr", opts.listen, "version", version)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func resolveToken(ctx context.Context, opts options, log *slog.Logger) (string, error) {
	if token := os.Getenv("FREEBOX_TOKEN"); token != "" {
		return token, nil
	}
	store := freebox.TokenStore{Path: filepath.Join(opts.dataDir, "token")}
	token, err := store.Load()
	if err != nil {
		return "", fmt.Errorf("load token: %w", err)
	}
	if token != "" {
		return token, nil
	}

	token, err = register(ctx, opts, log)
	if err != nil {
		return "", err
	}
	if err := store.Save(token); err != nil {
		return "", fmt.Errorf("save token in %s: %w", opts.dataDir, err)
	}
	return token, nil
}

func register(ctx context.Context, opts options, log *slog.Logger) (string, error) {
	client, err := freebox.New(ctx, opts.freeboxURL, appID, "", nil)
	if err != nil {
		return "", err
	}
	device, _ := os.Hostname()
	app := freebox.Application{ID: appID, Name: appName, Version: version, Device: device}

	for {
		token, trackID, err := client.Authorize(ctx, app)
		if err != nil {
			return "", err
		}
		log.Info("no token found: press the right arrow button on the Freebox front panel to authorize the exporter")

		err = client.WaitGranted(ctx, trackID, 2*time.Second)
		switch {
		case err == nil:
			log.Info("exporter authorized; grant it access to the settings in Freebox OS (Parameters > Access management > Applications)")
			return token, nil
		case errors.Is(err, freebox.ErrAuthorizationTimeout):
			log.Warn("authorization timed out, asking again")
		default:
			return "", err
		}
	}
}

func defaultDataDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "."
	}
	return filepath.Join(dir, appID)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
