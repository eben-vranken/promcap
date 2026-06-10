package promcap

import "github.com/prometheus/client_golang/prometheus"

type CappedHistogramVec struct {
	histogramVec *prometheus.HistogramVec
	lim          *limiter
}

func (c *Cap) NewHistogramVec(opts prometheus.HistogramOpts, labels []string, capOpts CapOpts) *CappedHistogramVec {
	hgv := prometheus.NewHistogramVec(opts, labels)
	c.reg.MustRegister(hgv)

	return &CappedHistogramVec{
		histogramVec: hgv,
		lim:          newLimiter(opts.Name, labels, capOpts, c.cappedTotal),
	}
}

func (hgv *CappedHistogramVec) WithLabelValues(lvs ...string) prometheus.Observer {
	return hgv.histogramVec.WithLabelValues(hgv.lim.resolve(lvs)...)
}

func (hgv *CappedHistogramVec) With(labels prometheus.Labels) prometheus.Observer {
	return hgv.histogramVec.WithLabelValues(hgv.lim.resolve(hgv.lim.order(labels))...)
}
