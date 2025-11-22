package syncer

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/tazhate/push-from-k8s-back-to-docker-registry/internal/config"
	"github.com/tazhate/push-from-k8s-back-to-docker-registry/internal/k8s"
	"github.com/tazhate/push-from-k8s-back-to-docker-registry/internal/metrics"
	"github.com/tazhate/push-from-k8s-back-to-docker-registry/internal/registry"
)

// Syncer manages the image synchronization process
type Syncer struct {
	config         *config.Config
	k8sClient      *k8s.Client
	registryClient *registry.Client
	logger         zerolog.Logger
}

// New creates a new Syncer instance
func New(cfg *config.Config, k8sClient *k8s.Client, registryClient *registry.Client, logger zerolog.Logger) *Syncer {
	return &Syncer{
		config:         cfg,
		k8sClient:      k8sClient,
		registryClient: registryClient,
		logger:         logger,
	}
}

// Run starts the synchronization loop
func (s *Syncer) Run(ctx context.Context) error {
	s.logger.Info().
		Dur("sync_period", s.config.SyncPeriod).
		Strs("namespaces", s.config.Namespaces).
		Msg("Starting image synchronization service")

	ticker := time.NewTicker(s.config.SyncPeriod)
	defer ticker.Stop()

	// Run initial sync immediately
	if err := s.syncOnce(ctx); err != nil {
		s.logger.Error().Err(err).Msg("Initial sync failed")
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info().Msg("Shutting down syncer")
			return ctx.Err()
		case <-ticker.C:
			if err := s.syncOnce(ctx); err != nil {
				s.logger.Error().Err(err).Msg("Sync cycle failed")
			}
		}
	}
}

// syncOnce performs a single synchronization cycle
func (s *Syncer) syncOnce(ctx context.Context) error {
	start := time.Now()
	defer func() {
		metrics.SyncDuration.Observe(time.Since(start).Seconds())
	}()

	s.logger.Info().Msg("Starting sync cycle")

	// Get all images from Kubernetes
	images, err := s.k8sClient.GetAllImages(ctx, s.config.Namespaces, s.config.Deployments)
	if err != nil {
		return err
	}

	if len(images) == 0 {
		s.logger.Warn().Msg("No images found to sync")
		return nil
	}

	metrics.ImagesProcessed.Set(float64(len(images)))

	s.logger.Info().
		Int("count", len(images)).
		Msg("Found images to process")

	// Process images with concurrency control
	s.syncImages(ctx, images)

	s.logger.Info().
		Dur("duration", time.Since(start)).
		Msg("Sync cycle completed")

	return nil
}

// syncImages syncs multiple images with concurrency control
func (s *Syncer) syncImages(ctx context.Context, images []string) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Max 5 concurrent syncs

	for _, image := range images {
		wg.Add(1)
		go func(img string) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				return
			}

			// Sync with retries
			if err := s.syncImageWithRetry(ctx, img); err != nil {
				s.logger.Error().
					Err(err).
					Str("image", img).
					Msg("Failed to sync image after retries")
			}
		}(image)
	}

	wg.Wait()
}

// syncImageWithRetry syncs a single image with retry logic
func (s *Syncer) syncImageWithRetry(ctx context.Context, image string) error {
	var lastErr error

	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			s.logger.Info().
				Str("image", image).
				Int("attempt", attempt).
				Int("max_retries", s.config.MaxRetries).
				Msg("Retrying image sync")

			select {
			case <-time.After(s.config.RetryDelay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := s.registryClient.SyncImage(ctx, image)
		if err == nil {
			return nil
		}

		lastErr = err
		s.logger.Warn().
			Err(err).
			Str("image", image).
			Int("attempt", attempt+1).
			Msg("Image sync attempt failed")
	}

	return lastErr
}
