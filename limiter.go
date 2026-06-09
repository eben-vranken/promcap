package promcap

import (
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type limiter struct {
	maxSeries int
	mu        sync.Mutex
	seen      map[string]struct{}
	name      string
	meta      *prometheus.CounterVec
}

func newLimiter(name string, maxSeries int, meta *prometheus.CounterVec) *limiter {
	return &limiter{
		maxSeries: maxSeries,
		seen:      make(map[string]struct{}),
		name:      name,
		meta:      meta,
	}
}

func (lim *limiter) resolve(lvs []string) []string {
	lim.mu.Lock()
	defer lim.mu.Unlock()

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
