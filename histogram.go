package promcap

import "github.com/prometheus/client_golang/prometheus"

type CappedHistogramVec struct {
	histogramVec *prometheus.HistogramVec
	lim          *limiter
}

func (c *Cap) NewHistogramVec(opts prometheus.HistogramOpts, labels []string, maxSeries int) *CappedHistogramVec {
	hgv := prometheus.NewHistogramVec(opts, labels)
	c.reg.MustRegister(hgv)

	return &CappedHistogramVec{
		histogramVec: hgv,
		lim:          newLimiter(maxSeries),
	}
}

func (hgv *CappedHistogramVec) WithLabelValues(lvs ...string) prometheus.Observer {
	return hgv.histogramVec.WithLabelValues(hgv.lim.resolve(lvs)...)
}
