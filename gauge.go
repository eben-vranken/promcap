package promcap

import "github.com/prometheus/client_golang/prometheus"

type CappedGaugeVec struct {
	gaugeVec *prometheus.GaugeVec
	lim      *limiter
}

func (c *Cap) NewGaugeVec(opts prometheus.GaugeOpts, labels []string, capOpts CapOpts) *CappedGaugeVec {
	gv := prometheus.NewGaugeVec(opts, labels)
	c.reg.MustRegister(gv)

	return &CappedGaugeVec{
		gaugeVec: gv,
		lim:      newLimiter(opts.Name, labels, capOpts, c.cappedTotal),
	}
}

func (cgv *CappedGaugeVec) WithLabelValues(lvs ...string) prometheus.Gauge {
	return cgv.gaugeVec.WithLabelValues(cgv.lim.resolve(lvs)...)
}

func (cgv *CappedGaugeVec) With(labels prometheus.Labels) prometheus.Gauge {
	return cgv.gaugeVec.WithLabelValues(cgv.lim.resolve(cgv.lim.order(labels))...)
}
