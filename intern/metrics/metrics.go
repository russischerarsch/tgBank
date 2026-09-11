package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Счётчик запросов (сколько всего пришло)
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// Гистограмма времени ответа
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	// Активные запросы сейчас
	HttpActiveRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_active_requests",
			Help: "Number of active HTTP requests",
		},
	)
)
var (
	// Счётчик запросов к БД
	DbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"query_type"}, // select / insert / update / delete
	)

	// Время выполнения запросов к БД
	DbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query latency",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"query_type"},
	)

	// Ошибки БД
	DbErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"query_type"},
	)
)
var (
	DbPoolConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_pool_connections",
			Help: "Number of database pool connections",
		},
		[]string{"state"}, // active / idle
	)
)

// ============================================================
// БИЗНЕС-МЕТРИКИ
// ============================================================

var (
	RentalsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "rentals_created_total",
			Help: "Total number of rentals created",
		},
	)

	RentalsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "rentals_active",
			Help: "Number of active rentals right now",
		},
	)

	CardsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cards_total",
			Help: "Total number of cards in the system",
		},
	)

	CardsFree = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cards_free",
			Help: "Number of free cards",
		},
	)

	CardsBusy = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cards_busy",
			Help: "Number of busy cards",
		},
	)

	UsersRegisteredTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "users_registered_total",
			Help: "Total number of registered users",
		},
	)
)
