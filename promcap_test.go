package promcap

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestSuccesfulPromcapInit(t *testing.T) {
	reg := prometheus.NewRegistry()
	regWrap := Wrap(reg)

	cv := regWrap.NewCounterVec(prometheus.CounterOpts{Name: "request_total"}, []string{"user"}, 2)
	cv.WithLabelValues("a").Inc()
	cv.WithLabelValues("b").Inc()
	cv.WithLabelValues("c").Inc()
	cv.WithLabelValues("d").Inc()

	count, err := testutil.GatherAndCount(reg)

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
