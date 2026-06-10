package promcap

import "github.com/prometheus/client_golang/prometheus"

type CappedGaugeVec struct {
	gaugeVec *prometheus.GaugeVec
	lim      *limiter
}

var _ prometheus.Collector = (*CappedGaugeVec)(nil)

func (c *Cap) NewGaugeVec(opts prometheus.GaugeOpts, labels []string, capOpts CapOpts) *CappedGaugeVec {
	gv := prometheus.NewGaugeVec(opts, labels)

	cgv := &CappedGaugeVec{
		gaugeVec: gv,
		lim:      newLimiter(opts.Name, labels, capOpts, c.cappedTotal),
	}

	c.reg.MustRegister(cgv)
	return cgv
}

func (cgv *CappedGaugeVec) WithLabelValues(lvs ...string) prometheus.Gauge {
	return cgv.gaugeVec.WithLabelValues(cgv.lim.resolve(lvs)...)
}

func (cgv *CappedGaugeVec) With(labels prometheus.Labels) prometheus.Gauge {
	return cgv.gaugeVec.WithLabelValues(cgv.lim.resolve(cgv.lim.order(labels))...)
}

func (cgv *CappedGaugeVec) Describe(ch chan<- *prometheus.Desc) { cgv.gaugeVec.Describe(ch) }
func (cgv *CappedGaugeVec) Collect(ch chan<- prometheus.Metric) { cgv.gaugeVec.Collect(ch) }

func (cgv *CappedGaugeVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Gauge, error) {
	return cgv.gaugeVec.GetMetricWithLabelValues(cgv.lim.resolve(lvs)...)
}

func (cgv *CappedGaugeVec) GetMetricWith(labels prometheus.Labels) (prometheus.Gauge, error) {
	return cgv.gaugeVec.GetMetricWithLabelValues(cgv.lim.resolve(cgv.lim.order(labels))...)
}

func (cgv *CappedGaugeVec) Reset() {
	cgv.gaugeVec.Reset()
	cgv.lim.reset()
}
