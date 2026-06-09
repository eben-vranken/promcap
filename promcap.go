package promcap

import (
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const overflowValue = "__overflow__"

type Cap struct {
	reg prometheus.Registerer
}

type limiter struct {
	maxSeries int
	mu        sync.Mutex
	seen      map[string]struct{}
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

func (c *Cap) NewCounterVec(opts prometheus.CounterOpts, labels []string, maxSeries int) *CappedCounterVec {
	cv := prometheus.NewCounterVec(opts, labels)
	c.reg.MustRegister(cv)

	return &CappedCounterVec{
		counterVec: cv,
		lim:        newLimiter(maxSeries),
	}
}

func Wrap(reg prometheus.Registerer) *Cap {
	return &Cap{reg: reg}
}

func newLimiter(maxSeries int) *limiter {
	return &limiter{
		maxSeries: maxSeries,
		seen:      make(map[string]struct{}),
	}
}

type CappedCounterVec struct {
	counterVec *prometheus.CounterVec
	lim        *limiter
}

func (ccv *CappedCounterVec) WithLabelValues(lvs ...string) prometheus.Counter {
	return ccv.counterVec.WithLabelValues(ccv.lim.resolve(lvs)...)
}
