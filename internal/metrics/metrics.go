// Package metrics holds Prometheus instruments that are used across the
// framework.  All collectors are registered with the global registry, so
// importing this package in main.go is enough to expose them on /metrics.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ActiveTenants = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_tenants",
			Help: "Number of tenants currently loaded in memory.",
		})

	TenantLoadTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "tenant_load_total",
			Help: "Cumulative number of tenants successfully loaded.",
		})

	TenantLoadErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "tenant_load_errors_total",
			Help: "Cumulative number of tenant load errors.",
		})

	TenantEvictTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "tenant_evict_total",
			Help: "Cumulative number of tenants evicted from the cache.",
		})
)

func init() {
	prometheus.MustRegister(
		ActiveTenants,
		TenantLoadTotal,
		TenantLoadErrorsTotal,
		TenantEvictTotal,
	)
}
