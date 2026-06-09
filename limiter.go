package promcap

import (
	"strings"
	"sync"
)

type limiter struct {
	maxSeries int
	mu        sync.Mutex
	seen      map[string]struct{}
}

func newLimiter(maxSeries int) *limiter {
	return &limiter{
		maxSeries: maxSeries,
		seen:      make(map[string]struct{}),
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

	overflow := make([]string, len(lvs))

	for i, _ := range overflow {
		overflow[i] = overflowValue
	}

	return overflow
}
