package metrics

import (
	"net/http"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// `method_duration` for latency of each call, in 99, 95, 50 quantiles
	RpcDurations = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_server_duration_seconds",
			Help:    "Histogram of gRPC request durations.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	// `method_count` to show count of failed/successful RPC calls
	RpcCalls = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_server_calls_total",
			Help: "Total number of gRPC calls.",
		},
		[]string{"code", "method"},
	)

	//`active_subscribers` to display total active subscriptions
	ActiveSubscriptions = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "grpc_active_subscriptions",
			Help: "Current number of active subscriptions.",
		},
	)

	MemStats = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "application_memory_usage_bytes",
			Help: "Current memory usage of the application in bytes.",
		},
		func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc)
		},
	)

	GcCount = prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "application_gc_count",
			Help: "Number of garbage collections performed by the application.",
		},
		func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.NumGC)
		},
	)

	CpuNum = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "application_cpu_num",
			Help: "Number of available cpus in the machine",
		},
		func() float64 {
			return float64(runtime.NumCPU())
		},
	)

	GoRoutineNum = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "application_goroutines_num",
			Help: "Number of active goroutines",
		},
		func() float64 {
			return float64(runtime.NumGoroutine())
		},
	)
)

func InitMetrics() {
	prometheus.MustRegister(RpcDurations)
	prometheus.MustRegister(RpcCalls)
	prometheus.MustRegister(ActiveSubscriptions)
	prometheus.MustRegister(MemStats)
	prometheus.MustRegister(GcCount)
	prometheus.MustRegister(CpuNum)
	prometheus.MustRegister(GoRoutineNum)
}

func StartMetricsServer() {
	InitMetrics()
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(":2112", nil); err != nil {
			panic(err)
		}
	}()
}
