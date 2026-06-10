package promcap

import "github.com/prometheus/client_golang/prometheus"

type CappedCounterVec struct {
	counterVec *prometheus.CounterVec
	lim        *limiter
}

func (c *Cap) NewCounterVec(opts prometheus.CounterOpts, labels []string, capOpts CapOpts) *CappedCounterVec {
	cv := prometheus.NewCounterVec(opts, labels)
	c.reg.MustRegister(cv)

	return &CappedCounterVec{
		counterVec: cv,
		lim:        newLimiter(opts.Name, labels, capOpts, c.cappedTotal),
	}
}

func (ccv *CappedCounterVec) WithLabelValues(lvs ...string) prometheus.Counter {
	return ccv.counterVec.WithLabelValues(ccv.lim.resolve(lvs)...)
}

func (ccv *CappedCounterVec) With(labels prometheus.Labels) prometheus.Counter {
	return ccv.counterVec.WithLabelValues(ccv.lim.resolve(ccv.lim.order(labels))...)
}
