package promcap

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestSuccesfulSummaryVecInit(t *testing.T) {
	reg := prometheus.NewRegistry()
	regWrap := Wrap(reg)

	cv := regWrap.NewSummaryVec(prometheus.SummaryOpts{Name: "request_total"}, []string{"user"}, CapOpts{MaxSeries: 2})
	cv.WithLabelValues("a").Observe(1)
	cv.WithLabelValues("b").Observe(1)
	cv.WithLabelValues("c").Observe(1)
	cv.WithLabelValues("d").Observe(1)

	count, err := testutil.GatherAndCount(reg, "request_total")

	if err != nil {
		t.Fatalf("Fatal error: %v", err)
	}

	if count != 3 {
		t.Errorf("Gather and count got %d, want %d", count, 3)
	}
}

func TestSummaryVecWith(t *testing.T) {
	reg := prometheus.NewRegistry()
	regWrap := Wrap(reg)

	cv := regWrap.NewSummaryVec(prometheus.SummaryOpts{Name: "request_total"}, []string{"user"}, CapOpts{MaxSeries: 1})
	cv.With(prometheus.Labels{"user": "a"}).Observe(1)
	cv.With(prometheus.Labels{"user": "b"}).Observe(1)

	count, err := testutil.GatherAndCount(reg, "request_total")

	if err != nil {
		t.Fatalf("Fatal error: %v", err)
	}

	if count != 2 {
		t.Errorf("Label A values got %v, want %v", count, 2)
	}

	if testutil.ToFloat64(regWrap.cappedTotal.WithLabelValues("request_total")) != 1 {
		t.Errorf("overflow total got %v, want %v",
			testutil.ToFloat64(regWrap.cappedTotal.WithLabelValues("request_total")), 1)
	}
}

func TestSummaryVecGetMetricWith(t *testing.T) {
	reg := prometheus.NewRegistry()
	regWrap := Wrap(reg)

	cv := regWrap.NewSummaryVec(prometheus.SummaryOpts{Name: "request_total"}, []string{"user"}, CapOpts{MaxSeries: 1})

	a, err := cv.GetMetricWith(prometheus.Labels{"user": "a"})
	if err != nil {
		t.Fatalf("GetMetricWith returned error: %v", err)
	}
	a.Observe(1)

	b, err := cv.GetMetricWithLabelValues("b")
	if err != nil {
		t.Fatalf("GetMetricWithLabelValues returned error: %v", err)
	}
	b.Observe(1)

	if testutil.ToFloat64(regWrap.cappedTotal.WithLabelValues("request_total")) != 1 {
		t.Errorf("overflow got %v, want %v", testutil.ToFloat64(regWrap.cappedTotal.WithLabelValues("request_total")), 1)
	}
}

func TestSummaryVecResetFreesBudget(t *testing.T) {
	reg := prometheus.NewRegistry()
	regWrap := Wrap(reg)

	cv := regWrap.NewSummaryVec(prometheus.SummaryOpts{Name: "request_total"}, []string{"user"}, CapOpts{MaxSeries: 1})
	cv.WithLabelValues("a").Observe(1)
	cv.WithLabelValues("b").Observe(1)

	cv.Reset()

	cv.WithLabelValues("c").Observe(1)

	count, err := testutil.GatherAndCount(reg, "request_total")

	if err != nil {
		t.Fatalf("Fatal error: %v", err)
	}

	if count != 1 {
		t.Errorf("Gather and Count got %v, want %v", count, 1)
	}

	if testutil.ToFloat64(regWrap.cappedTotal.WithLabelValues("request_total")) != 1 {
		t.Errorf("overflow total got %v, want %v",
			testutil.ToFloat64(regWrap.cappedTotal.WithLabelValues("request_total")), 1)
	}
}

func TestSummaryVecEvictionDeletesSeries(t *testing.T) {
	reg := prometheus.NewRegistry()
	regWrap := Wrap(reg)

	cv := regWrap.NewSummaryVec(prometheus.SummaryOpts{Name: "request_total"}, []string{"user"}, CapOpts{MaxSeries: 1, Evict: true})
	cv.WithLabelValues("a").Observe(1)
	cv.WithLabelValues("b").Observe(1)

	count, err := testutil.GatherAndCount(reg, "request_total")
	if err != nil {
		t.Fatalf("Fatal error: %v", err)
	}

	if count != 1 {
		t.Errorf("series after eviction got %d, want %d", count, 1)
	}

	if testutil.ToFloat64(regWrap.cappedTotal.WithLabelValues("request_total")) != 0 {
		t.Errorf("overflow total got %v, want %v",
			testutil.ToFloat64(regWrap.cappedTotal.WithLabelValues("request_total")), 0)
	}
}
