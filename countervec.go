package promcap

import "github.com/prometheus/client_golang/prometheus"

type CappedCounterVec struct {
	counterVec *prometheus.CounterVec
	lim        *limiter
}

func (c *Cap) NewCounterVec(opts prometheus.CounterOpts, labels []string, maxSeries int) *CappedCounterVec {
	cv := prometheus.NewCounterVec(opts, labels)
	c.reg.MustRegister(cv)

	return &CappedCounterVec{
		counterVec: cv,
		lim:        newLimiter(opts.Name, maxSeries, c.cappedTotal),
	}
}

func (ccv *CappedCounterVec) WithLabelValues(lvs ...string) prometheus.Counter {
	return ccv.counterVec.WithLabelValues(ccv.lim.resolve(lvs)...)
}
