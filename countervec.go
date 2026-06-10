package promcap

import "github.com/prometheus/client_golang/prometheus"

type CappedCounterVec struct {
	counterVec *prometheus.CounterVec
	lim        *limiter
}

var _ prometheus.Collector = (*CappedCounterVec)(nil)

func (c *Cap) NewCounterVec(opts prometheus.CounterOpts, labels []string, capOpts CapOpts) *CappedCounterVec {
	cv := prometheus.NewCounterVec(opts, labels)
	ccv := &CappedCounterVec{
		counterVec: cv,
		lim:        newLimiter(opts.Name, labels, capOpts, c.cappedTotal),
	}

	c.reg.MustRegister(ccv)

	return ccv
}

func (ccv *CappedCounterVec) WithLabelValues(lvs ...string) prometheus.Counter {
	return ccv.counterVec.WithLabelValues(ccv.lim.resolve(lvs)...)
}

func (ccv *CappedCounterVec) With(labels prometheus.Labels) prometheus.Counter {
	return ccv.counterVec.WithLabelValues(ccv.lim.resolve(ccv.lim.order(labels))...)
}

func (ccv *CappedCounterVec) Describe(ch chan<- *prometheus.Desc) { ccv.counterVec.Describe(ch) }
func (ccv *CappedCounterVec) Collect(ch chan<- prometheus.Metric) { ccv.counterVec.Collect(ch) }
