package promcap

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestSuccesfulHistogramVecInit(t *testing.T) {
	reg := prometheus.NewRegistry()
	regWrap := Wrap(reg)

	cv := regWrap.NewHistogramVec(prometheus.HistogramOpts{Name: "request_total"}, []string{"user"}, 2)
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
