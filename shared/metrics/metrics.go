package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// JobsReceivedTotal counts every job accepted at the API upload endpoint.
	JobsReceivedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "jobs_received_total",
		Help: "Total number of jobs received by the API.",
	}, []string{"job_type", "tenant_id"})

	// JobsCompletedTotal counts jobs that reached the completed state.
	JobsCompletedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "jobs_completed_total",
		Help: "Total number of jobs successfully completed.",
	}, []string{"job_type", "tenant_id"})

	// JobsFailedTotal counts jobs that reached a terminal failure state.
	JobsFailedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "jobs_failed_total",
		Help: "Total number of jobs that failed permanently.",
	}, []string{"job_type", "reason"})

	// JobDuration tracks end-to-end processing time per job type.
	JobDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "job_processing_duration_seconds",
		Help:    "Job processing time from SQS receipt to completion or failure.",
		Buckets: []float64{1, 5, 10, 30, 60, 120},
	}, []string{"job_type"})

	// SQSMessagesPolled counts raw SQS messages received per service.
	SQSMessagesPolled = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sqs_messages_polled_total",
		Help: "Total number of SQS messages polled.",
	}, []string{"service"})

	// ActiveSSEConnections is the current number of open SSE streams.
	ActiveSSEConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "active_sse_connections",
		Help: "Number of currently open SSE /events connections.",
	})
)
