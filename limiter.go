package promcap

import (
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type limiter struct {
	maxSeries  int
	mu         sync.Mutex
	seen       map[string]struct{}
	name       string
	meta       *prometheus.CounterVec
	labelNames []string
	allow      map[string]map[string]struct{}
}

func newLimiter(name string, labelNames []string, opts CapOpts, meta *prometheus.CounterVec) *limiter {
	allowSet := make(map[string]map[string]struct{})

	for label, values := range opts.Allow {
		permitted := make(map[string]struct{})

		for _, value := range values {
			permitted[value] = struct{}{}
		}

		allowSet[label] = permitted
	}

	return &limiter{
		maxSeries:  opts.MaxSeries,
		seen:       make(map[string]struct{}),
		name:       name,
		meta:       meta,
		labelNames: labelNames,
		allow:      allowSet,
	}
}

func (lim *limiter) resolve(lvs []string) []string {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	for i, name := range lim.labelNames {
		set, ok := lim.allow[name]

		if !ok {
			continue
		}

		_, allowed := set[lvs[i]]

		if !allowed {
			return lim.overflow(len(lvs))
		}
	}

	key := strings.Join(lvs, "\xff")

	_, ok := lim.seen[key]

	if ok {
		return lvs
	}

	if len(lim.seen) < lim.maxSeries {
		lim.seen[key] = struct{}{}
		return lvs
	}

	return lim.overflow(len(lvs))
}

func (lim *limiter) overflow(n int) []string {
	lim.meta.WithLabelValues(lim.name).Inc()

	overflow := make([]string, n)

	for i := range overflow {
		overflow[i] = overflowValue
	}

	return overflow
}
