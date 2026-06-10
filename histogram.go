package promcap

import "github.com/prometheus/client_golang/prometheus"

type CappedHistogramVec struct {
	histogramVec *prometheus.HistogramVec
	lim          *limiter
}

var _ prometheus.Collector = (*CappedHistogramVec)(nil)

func (c *Cap) NewHistogramVec(opts prometheus.HistogramOpts, labels []string, capOpts CapOpts) *CappedHistogramVec {
	hgv := prometheus.NewHistogramVec(opts, labels)

	chgv := &CappedHistogramVec{
		histogramVec: hgv,
		lim:          newLimiter(opts.Name, labels, capOpts, c.cappedTotal),
	}

	c.reg.MustRegister(chgv)

	return chgv
}

func (hgv *CappedHistogramVec) WithLabelValues(lvs ...string) prometheus.Observer {
	return hgv.histogramVec.WithLabelValues(hgv.lim.resolve(lvs)...)
}

func (hgv *CappedHistogramVec) With(labels prometheus.Labels) prometheus.Observer {
	return hgv.histogramVec.WithLabelValues(hgv.lim.resolve(hgv.lim.order(labels))...)
}

func (hgv *CappedHistogramVec) Describe(ch chan<- *prometheus.Desc) { hgv.histogramVec.Describe(ch) }
func (hgv *CappedHistogramVec) Collect(ch chan<- prometheus.Metric) { hgv.histogramVec.Collect(ch) }
