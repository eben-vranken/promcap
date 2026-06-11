package promcap

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestSuccesfulGaugeVecInit(t *testing.T) {
	reg := prometheus.NewRegistry()
	regWrap := Wrap(reg)

	cv := regWrap.NewGaugeVec(prometheus.GaugeOpts{Name: "request_total"}, []string{"user"}, CapOpts{MaxSeries: 2})
	cv.WithLabelValues("a").Inc()
	cv.WithLabelValues("b").Inc()
	cv.WithLabelValues("c").Inc()
	cv.WithLabelValues("d").Inc()

	count, err := testutil.GatherAndCount(reg, "request_total")

	if err != nil {
		t.Fatalf("Fatal error: %v", err)
	}

	if count != 3 {
		t.Errorf("Gather and count got %d, want %d", count, 3)
	}

	if testutil.ToFloat64(cv.WithLabelValues(overflowValue)) != 2 {
		t.Errorf("Overflow values got %v, want %v", testutil.ToFloat64(cv.WithLabelValues(overflowValue)), 2)
	}
}

func TestGaugeVecWith(t *testing.T) {
	reg := prometheus.NewRegistry()
	regWrap := Wrap(reg)

	cv := regWrap.NewGaugeVec(prometheus.GaugeOpts{Name: "request_total"}, []string{"user"}, CapOpts{MaxSeries: 1})
	cv.With(prometheus.Labels{"user": "a"}).Inc()
	cv.With(prometheus.Labels{"user": "b"}).Inc()

	if testutil.ToFloat64(cv.With(prometheus.Labels{"user": "a"})) != 1 {
		t.Errorf("Label A values got %v, want %v", testutil.ToFloat64(cv.With(prometheus.Labels{"user": "a"})), 1)
	}

	if testutil.ToFloat64(cv.WithLabelValues(overflowValue)) != 1 {
		t.Errorf("Overflow values got %v, want %v", testutil.ToFloat64(cv.WithLabelValues(overflowValue)), 1)
	}
}

func TestGaugeVecGetMetricWith(t *testing.T) {
	reg := prometheus.NewRegistry()
	regWrap := Wrap(reg)

	cv := regWrap.NewGaugeVec(prometheus.GaugeOpts{Name: "request_total"}, []string{"user"}, CapOpts{MaxSeries: 1})

	a, err := cv.GetMetricWith(prometheus.Labels{"user": "a"})
	if err != nil {
		t.Fatalf("GetMetricWith returned error: %v", err)
	}
	a.Inc()

	b, err := cv.GetMetricWithLabelValues("b")
	if err != nil {
		t.Fatalf("GetMetricWithLabelValues returned error: %v", err)
	}
	b.Inc()

	if testutil.ToFloat64(a) != 1 {
		t.Errorf("admitted series got %v, want %v", testutil.ToFloat64(a), 1)
	}

	if testutil.ToFloat64(cv.WithLabelValues(overflowValue)) != 1 {
		t.Errorf("overflow got %v, want %v", testutil.ToFloat64(cv.WithLabelValues(overflowValue)), 1)
	}
}

func TestGaugeVecResetFreesBudget(t *testing.T) {
	reg := prometheus.NewRegistry()
	regWrap := Wrap(reg)

	cv := regWrap.NewGaugeVec(prometheus.GaugeOpts{Name: "request_total"}, []string{"user"}, CapOpts{MaxSeries: 1})
	cv.WithLabelValues("a").Inc()
	cv.WithLabelValues("b").Inc()

	cv.Reset()

	cv.WithLabelValues("c").Inc()

	if testutil.ToFloat64(cv.WithLabelValues("c")) != 1 {
		t.Errorf("Label C values got %v, want %v", testutil.ToFloat64(cv.With(prometheus.Labels{"user": "c"})), 1)
	}

	if testutil.ToFloat64(regWrap.cappedTotal.WithLabelValues("request_total")) != 1 {
		t.Errorf("overflow total got %v, want %v",
			testutil.ToFloat64(regWrap.cappedTotal.WithLabelValues("request_total")), 1)
	}
}
