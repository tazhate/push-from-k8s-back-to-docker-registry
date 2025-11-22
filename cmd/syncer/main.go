package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tazhate/push-from-k8s-back-to-docker-registry/internal/config"
	"github.com/tazhate/push-from-k8s-back-to-docker-registry/internal/k8s"
	"github.com/tazhate/push-from-k8s-back-to-docker-registry/internal/registry"
	"github.com/tazhate/push-from-k8s-back-to-docker-registry/internal/syncer"
)

func main() {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Set log level
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.Warn().Str("level", cfg.LogLevel).Msg("Invalid log level, using info")
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	logger.Info().
		Str("registry", cfg.RegistryURL).
		Strs("namespaces", cfg.Namespaces).
		Dur("sync_period", cfg.SyncPeriod).
		Msg("Starting image sync service")

	// Create Kubernetes client
	k8sClient, err := k8s.NewClient(logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}
	logger.Info().Msg("Kubernetes client initialized")

	// Detect container runtime socket first (before creating syncer)
	containerdSocketPath, runtimeType, err := registry.DetectContainerdSocket(cfg.ContainerdSocketPath, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to detect container runtime socket")
	}

	// Create registry client
	registryClient, err := registry.NewClient(cfg.RegistryURL, cfg.RegistryUsername, cfg.RegistryPassword, containerdSocketPath, runtimeType, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create registry client")
	}
	logger.Info().Str("runtime", string(runtimeType)).Msg("Registry client initialized")

	// Create syncer
	syncerInstance := syncer.New(cfg, k8sClient, registryClient, logger)

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start metrics server
	go startMetricsServer(cfg.MetricsAddr, logger)

	// Start health server
	go startHealthServer(cfg.HealthAddr, logger)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start syncer in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- syncerInstance.Run(ctx)
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		logger.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
		cancel()
		// Give some time for graceful shutdown
		time.Sleep(2 * time.Second)
	case err := <-errChan:
		if err != nil && err != context.Canceled {
			logger.Error().Err(err).Msg("Syncer error")
			os.Exit(1)
		}
	}

	logger.Info().Msg("Shutdown complete")
}

// startMetricsServer starts the Prometheus metrics HTTP server
func startMetricsServer(addr string, logger zerolog.Logger) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	logger.Info().Str("addr", addr).Msg("Starting metrics server")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error().Err(err).Msg("Metrics server error")
	}
}

// startHealthServer starts the health check HTTP server
func startHealthServer(addr string, logger zerolog.Logger) {
	mux := http.NewServeMux()

	// Liveness probe
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	// Readiness probe
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Ready")
	})

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	logger.Info().Str("addr", addr).Msg("Starting health server")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error().Err(err).Msg("Health server error")
	}
}
