package promcap

import "github.com/prometheus/client_golang/prometheus"

type CappedGaugeVec struct {
	gaugeVec *prometheus.GaugeVec
	lim      *limiter
}

func (c *Cap) NewGaugeVec(opts prometheus.GaugeOpts, labels []string, maxSeries int) *CappedGaugeVec {
	gv := prometheus.NewGaugeVec(opts, labels)
	c.reg.MustRegister(gv)

	return &CappedGaugeVec{
		gaugeVec: gv,
		lim:      newLimiter(opts.Name, maxSeries, c.cappedTotal),
	}
}

func (cgv *CappedGaugeVec) WithLabelValues(lvs ...string) prometheus.Gauge {
	return cgv.gaugeVec.WithLabelValues(cgv.lim.resolve(lvs)...)
}
