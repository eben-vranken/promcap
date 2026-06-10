package promcap

import "github.com/prometheus/client_golang/prometheus"

type CappedSummaryVec struct {
	summaryVec *prometheus.SummaryVec
	lim        *limiter
}

var _ prometheus.Collector = (*CappedSummaryVec)(nil)

func (c *Cap) NewSummaryVec(opts prometheus.SummaryOpts, labels []string, capOpts CapOpts) *CappedSummaryVec {
	sumv := prometheus.NewSummaryVec(opts, labels)

	csumv := &CappedSummaryVec{
		summaryVec: sumv,
		lim:        newLimiter(opts.Name, labels, capOpts, c.cappedTotal),
	}
	csumv.lim.onEvict = func(lvs []string) { sumv.DeleteLabelValues(lvs...) }

	c.reg.MustRegister(csumv)
	return csumv
}

func (sumv *CappedSummaryVec) WithLabelValues(lvs ...string) prometheus.Observer {
	return sumv.summaryVec.WithLabelValues(sumv.lim.resolve(lvs)...)
}

func (sumv *CappedSummaryVec) With(labels prometheus.Labels) prometheus.Observer {
	return sumv.summaryVec.WithLabelValues(sumv.lim.resolve(sumv.lim.order(labels))...)
}

func (sumv *CappedSummaryVec) Describe(ch chan<- *prometheus.Desc) { sumv.summaryVec.Describe(ch) }
func (sumv *CappedSummaryVec) Collect(ch chan<- prometheus.Metric) { sumv.summaryVec.Collect(ch) }

func (sumv *CappedSummaryVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Observer, error) {
	return sumv.summaryVec.GetMetricWithLabelValues(sumv.lim.resolve(lvs)...)
}

func (sumv *CappedSummaryVec) GetMetricWith(labels prometheus.Labels) (prometheus.Observer, error) {
	return sumv.summaryVec.GetMetricWithLabelValues(sumv.lim.resolve(sumv.lim.order(labels))...)
}

func (sumv *CappedSummaryVec) Reset() {
	sumv.summaryVec.Reset()
	sumv.lim.reset()
}
