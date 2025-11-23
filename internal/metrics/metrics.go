package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ImagesSynced tracks successfully synced images
	ImagesSynced = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "images_synced_total",
			Help: "Total number of images successfully synced to target registry",
		},
		[]string{"source_registry", "target_registry"},
	)

	// ImagesSyncFailed tracks failed sync attempts
	ImagesSyncFailed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "images_sync_failed_total",
			Help: "Total number of images that failed to sync",
		},
		[]string{"source_registry", "target_registry", "reason"},
	)

	// ImagesSkipped tracks images skipped because they already exist
	ImagesSkipped = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "images_skipped_total",
			Help: "Total number of images skipped (already exist in target)",
		},
		[]string{"source_registry", "target_registry"},
	)

	// SyncDuration tracks sync cycle duration
	SyncDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "sync_cycle_duration_seconds",
			Help:    "Duration of sync cycles in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	// ImagesProcessed tracks total images processed per cycle
	ImagesProcessed = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "images_processed_current",
			Help: "Number of images processed in current sync cycle",
		},
	)

	// RegistryAuthentications tracks registry auth attempts
	RegistryAuthentications = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "registry_authentications_total",
			Help: "Total number of registry authentication attempts",
		},
		[]string{"registry", "status"},
	)

	// ImageSyncDuration tracks individual image sync duration
	ImageSyncDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "image_sync_duration_seconds",
			Help:    "Duration of individual image sync operations",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
		},
		[]string{"operation"},
	)
)
