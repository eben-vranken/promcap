// Package promcap wraps Prometheus *Vec metrics with a cardinality cap,
// collapsing label combinations beyond a configured limit into an overflow series.
//
// The label value "__overflow__" is reserved: any real label value equal to it
// will silently merge into the overflow bucket.
package promcap

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

// overflowValue is the reserved label value that collapsed series carry.
// A real label value equal to it will silently merge into the overflow bucket.
const overflowValue = "__overflow__"

type Cap struct {
	reg         prometheus.Registerer
	cappedTotal *prometheus.CounterVec
}

func Wrap(reg prometheus.Registerer) *Cap {
	c := &Cap{
		reg:         reg,
		cappedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "promcap_series_capped_total", Help: "Total number of observations collapsed in the overflow series, by metric"}, []string{"metric"}),
	}

	var are prometheus.AlreadyRegisteredError
	err := reg.Register(c.cappedTotal)

	if errors.As(err, &are) {
		c.cappedTotal = are.ExistingCollector.(*prometheus.CounterVec)
	} else if err != nil {
		panic(err)
	}

	return c
}
