package promcap

import (
	"fmt"
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
	lookupSet := make(map[string]struct{})

	for _, label := range labelNames {
		lookupSet[label] = struct{}{}
	}

	for label, values := range opts.Allow {
		_, ok := lookupSet[label]

		if !ok {
			panic(fmt.Sprintf("promcap: Allow key %q is not a label of metric %q", label, name))
		}

		permitted := make(map[string]struct{})

		for _, value := range values {
			permitted[value] = struct{}{}
		}

		allowSet[label] = permitted
	}

	if opts.MaxSeries <= 0 {
		opts.MaxSeries = defaultMaxSeries
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
			return lim.overflow(lvs)
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

	return lim.overflow(lvs)
}

func (lim *limiter) overflow(lvs []string) []string {
	lim.meta.WithLabelValues(lim.name).Inc()

	out := make([]string, len(lvs))
	collapsedAny := false

	for i, name := range lim.labelNames {
		set, ok := lim.allow[name]

		if ok {
			_, allowed := set[lvs[i]]

			if allowed {
				out[i] = lvs[i]
				continue
			}
		}

		out[i] = overflowValue
		collapsedAny = true
	}

	if !collapsedAny {
		for i := range out {
			out[i] = overflowValue
		}
	}

	return out
}

func (lim *limiter) order(labels prometheus.Labels) []string {
	lvs := make([]string, len(lim.labelNames))
	for i, name := range lim.labelNames {
		v, ok := labels[name]
		if !ok {
			panic(fmt.Sprintf("promcap: missing label %q for metric %q", name, lim.name))
		}
		lvs[i] = v
	}
	return lvs
}

func (lim *limiter) reset() {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	lim.seen = make(map[string]struct{})
}
