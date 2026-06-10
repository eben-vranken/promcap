package promcap

import (
	"container/list"
	"fmt"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type limiter struct {
	maxSeries   int
	mu          sync.Mutex
	seen        map[string]*list.Element
	name        string
	metaCounter prometheus.Counter
	labelNames  []string
	allow       map[string]map[string]struct{}
	lru         *list.List
	evict       bool
	onEvict     func(lvs []string)
}

type lruEntry struct {
	key string
	lvs []string
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
		maxSeries:   opts.MaxSeries,
		seen:        make(map[string]*list.Element),
		name:        name,
		metaCounter: meta.WithLabelValues(name),
		labelNames:  labelNames,
		allow:       allowSet,
		lru:         list.New(),
		evict:       opts.Evict,
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

	if elem, ok := lim.seen[key]; ok {
		lim.lru.MoveToFront(elem)
		return lvs
	}

	if len(lim.seen) >= lim.maxSeries {
		if !lim.evict {
			return lim.overflow(lvs)
		}
		lim.evictOldest()
	}

	elem := lim.lru.PushFront(lruEntry{key: key, lvs: append([]string(nil), lvs...)})
	lim.seen[key] = elem
	return lvs

}

func (lim *limiter) overflow(lvs []string) []string {
	lim.metaCounter.Inc()

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
	lim.seen = make(map[string]*list.Element)
	lim.lru = list.New()
}

func (lim *limiter) evictOldest() {
	back := lim.lru.Back()
	if back == nil {
		return
	}
	ent := back.Value.(lruEntry)
	lim.lru.Remove(back)
	delete(lim.seen, ent.key)
	if lim.onEvict != nil {
		lim.onEvict(ent.lvs)
	}
}
