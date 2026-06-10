package promcap

import "github.com/prometheus/client_golang/prometheus"

type CappedSummaryVec struct {
	summaryVec *prometheus.SummaryVec
	lim        *limiter
}

func (c *Cap) NewSummaryVec(opts prometheus.SummaryOpts, labels []string, capOpts CapOpts) *CappedSummaryVec {
	sumv := prometheus.NewSummaryVec(opts, labels)
	c.reg.MustRegister(sumv)

	return &CappedSummaryVec{
		summaryVec: sumv,
		lim:        newLimiter(opts.Name, labels, capOpts, c.cappedTotal),
	}
}

func (sumv *CappedSummaryVec) WithLabelValues(lvs ...string) prometheus.Observer {
	return sumv.summaryVec.WithLabelValues(sumv.lim.resolve(lvs)...)
}

func (sumv *CappedSummaryVec) With(labels prometheus.Labels) prometheus.Observer {
	return sumv.summaryVec.WithLabelValues(sumv.lim.resolve(sumv.lim.order(labels))...)
}
