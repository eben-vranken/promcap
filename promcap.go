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

func (c *Cap) NewCounterVec(opts prometheus.CounterOpts, labels []string, maxSeries int) *CappedCounterVec {
	cv := prometheus.NewCounterVec(opts, labels)
	c.reg.MustRegister(cv)

	return &CappedCounterVec{
		counterVec: cv,
		maxSeries:  maxSeries,
		seen:       make(map[string]struct{}),
	}
}

func Wrap(reg prometheus.Registerer) *Cap {
	return &Cap{reg: reg}
}

type CappedCounterVec struct {
	counterVec *prometheus.CounterVec
	maxSeries  int
	mu         sync.Mutex
	seen       map[string]struct{}
}

func (ccv *CappedCounterVec) WithLabelValues(lvs ...string) prometheus.Counter {
	ccv.mu.Lock()
	defer ccv.mu.Unlock()

	key := strings.Join(lvs, "\xff")

	_, ok := ccv.seen[key]

	if ok {
		return ccv.counterVec.WithLabelValues(lvs...)
	}

	if len(ccv.seen) < ccv.maxSeries {
		ccv.seen[key] = struct{}{}
		return ccv.counterVec.WithLabelValues(lvs...)
	}

	overflow := make([]string, len(lvs))

	for i, _ := range overflow {
		overflow[i] = overflowValue
	}

	return ccv.counterVec.WithLabelValues(overflow...)
}
