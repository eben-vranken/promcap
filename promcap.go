package promcap

import (
	"github.com/prometheus/client_golang/prometheus"
)

const overflowValue = "__overflow__"

type Cap struct {
	reg         prometheus.Registerer
	cappedTotal *prometheus.CounterVec
}

func Wrap(reg prometheus.Registerer) *Cap {
	c := &Cap{
		reg:         reg,
		cappedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "promcap_series_capped_total", Help: "Number of label combinations collapsed into the overflow series, by metric"}, []string{"metric"}),
	}

	reg.MustRegister(c.cappedTotal)

	return c
}
