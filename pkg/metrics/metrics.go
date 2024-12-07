package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ConfigMapOperations tracks operations performed on ConfigMaps
	ConfigMapOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "configmap_operator_operations_total",
			Help: "The total number of ConfigMap operations performed by the operator",
		},
		[]string{"operation", "status"}, // operation: create, update, delete; status: success, error
	)

	// ConfigMapProcessingLatency tracks the time taken to process ConfigMap operations
	ConfigMapProcessingLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "configmap_operator_processing_duration_seconds",
			Help:    "Time taken to process ConfigMap operations",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // start at 1ms, double 10 times
		},
		[]string{"operation"},
	)

	// ConfigMapErrors tracks specific error types
	ConfigMapErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "configmap_operator_errors_total",
			Help: "The total number of errors encountered by the operator",
		},
		[]string{"error_type"}, // error_type: validation, kubernetes_api, merge
	)

	// MergedConfigMapsSize tracks the size of merged ConfigMaps
	MergedConfigMapsSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "configmap_operator_merged_size_bytes",
			Help: "Size of merged ConfigMaps in bytes",
		},
		[]string{"namespace", "name"},
	)

	// WatchedConfigMapsCount tracks the number of ConfigMaps being watched
	WatchedConfigMapsCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "configmap_operator_watched_configmaps",
			Help: "Number of ConfigMaps being watched by the operator",
		},
		[]string{"namespace"},
	)
)
